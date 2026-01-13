use notify::{Config, RecommendedWatcher, RecursiveMode, Watcher, Event, EventKind};
use serde::{Deserialize, Serialize};
use std::collections::HashMap;
use std::fs;
use std::path::PathBuf;
use std::sync::{Arc, Mutex};
use std::time::{Duration, Instant};
use tauri::{AppHandle, Emitter};
use uuid::Uuid;
use chrono::Local;

const UPLOAD_COMPLETE_TIMEOUT: u64 = 30;

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppConfig {
    pub watch_folder: String,
    pub file_types: Vec<String>,
    pub watch_subdirs: bool,
    pub sound_enabled: bool,
    pub ignore_folders: Vec<String>,
    pub save_history: bool,
}

impl Default for AppConfig {
    fn default() -> Self {
        Self {
            watch_folder: String::new(),
            file_types: vec![
                ".mp4".into(), ".avi".into(), ".mkv".into(), ".mov".into(),
                ".wmv".into(), ".flv".into(), ".webm".into(), ".m4v".into(),
                ".mpeg".into(), ".mpg".into(), ".3gp".into(), ".ts".into(),
            ],
            watch_subdirs: true,
            sound_enabled: true,
            ignore_folders: vec![
                "node_modules".into(), ".git".into(), "__pycache__".into(),
                ".idea".into(), "vendor".into(), "target".into(),
            ],
            save_history: true,
        }
    }
}

#[derive(Debug, Clone, Serialize, Deserialize, PartialEq)]
pub enum BatchStatus {
    Uploading,
    Completed,
    Signed,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct Batch {
    pub id: String,
    pub folder: String,
    pub files: Vec<String>,
    pub status: BatchStatus,
    pub started_at: String,
    pub completed_at: Option<String>,
    pub signed_at: Option<String>,
    pub file_count: usize,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct AppState {
    pub is_running: bool,
    pub batches: Vec<Batch>,
    pub unsigned_count: i32,
    pub uploading_count: i32,
    pub dir_count: i32,
    pub start_time: String,
}

struct FolderTracker {
    files: Vec<String>,
    last_activity: Instant,
    started_at: String,
    notified_start: bool,
    batch_id: String,
}

pub struct GlobalState {
    pub config: AppConfig,
    pub batches: Vec<Batch>,
    pub folder_trackers: HashMap<String, FolderTracker>,
    pub is_running: bool,
    pub dir_count: i32,
    pub start_time: String,
    pub watcher: Option<RecommendedWatcher>,
}

impl Default for GlobalState {
    fn default() -> Self {
        Self {
            config: AppConfig::default(),
            batches: Vec::new(),
            folder_trackers: HashMap::new(),
            is_running: false,
            dir_count: 0,
            start_time: String::new(),
            watcher: None,
        }
    }
}

type SharedState = Arc<Mutex<GlobalState>>;

fn get_config_path() -> PathBuf {
    let config_dir = dirs::config_dir()
        .unwrap_or_else(|| PathBuf::from("."))
        .join("fidruawatch");
    fs::create_dir_all(&config_dir).ok();
    config_dir.join("config.json")
}

fn get_history_path() -> PathBuf {
    let config_dir = dirs::config_dir()
        .unwrap_or_else(|| PathBuf::from("."))
        .join("fidruawatch");
    fs::create_dir_all(&config_dir).ok();
    config_dir.join("history.json")
}

fn load_config() -> AppConfig {
    let path = get_config_path();
    fs::read_to_string(&path)
        .ok()
        .and_then(|data| serde_json::from_str(&data).ok())
        .unwrap_or_default()
}

fn save_config(config: &AppConfig) {
    if let Ok(data) = serde_json::to_string_pretty(config) {
        fs::write(get_config_path(), data).ok();
    }
}

fn load_history() -> Vec<Batch> {
    fs::read_to_string(get_history_path())
        .ok()
        .and_then(|data| serde_json::from_str(&data).ok())
        .unwrap_or_default()
}

fn save_history(batches: &[Batch]) {
    if let Ok(data) = serde_json::to_string_pretty(batches) {
        fs::write(get_history_path(), data).ok();
    }
}

#[tauri::command]
fn get_config(state: tauri::State<SharedState>) -> AppConfig {
    state.lock().unwrap().config.clone()
}

#[tauri::command]
fn set_config(state: tauri::State<SharedState>, config: AppConfig) {
    let mut s = state.lock().unwrap();
    s.config = config.clone();
    save_config(&config);
}

#[tauri::command]
fn get_state(state: tauri::State<SharedState>) -> AppState {
    let s = state.lock().unwrap();
    AppState {
        is_running: s.is_running,
        batches: s.batches.clone(),
        unsigned_count: s.batches.iter().filter(|b| b.status == BatchStatus::Completed).count() as i32,
        uploading_count: s.batches.iter().filter(|b| b.status == BatchStatus::Uploading).count() as i32,
        dir_count: s.dir_count,
        start_time: s.start_time.clone(),
    }
}

#[tauri::command]
fn start_monitor(app: AppHandle, state: tauri::State<SharedState>) -> bool {
    let mut s = state.lock().unwrap();
    
    if s.is_running { return false; }
    
    let watch_folder = s.config.watch_folder.clone();
    if watch_folder.is_empty() || !PathBuf::from(&watch_folder).exists() {
        return false;
    }
    
    let config = s.config.clone();
    let state_clone = Arc::clone(&state.inner());
    let app_handle = app.clone();
    
    let watcher = RecommendedWatcher::new(
        move |res: Result<Event, notify::Error>| {
            if let Ok(event) = res {
                handle_file_event(event, &state_clone, &config, &app_handle);
            }
        },
        Config::default(),
    );
    
    match watcher {
        Ok(mut w) => {
            let mode = if s.config.watch_subdirs { RecursiveMode::Recursive } else { RecursiveMode::NonRecursive };
            if w.watch(watch_folder.as_ref(), mode).is_err() { return false; }
            
            s.watcher = Some(w);
            s.is_running = true;
            s.start_time = Local::now().format("%H:%M:%S").to_string();
            s.dir_count = 1;
            
            let state_for_checker = Arc::clone(&state.inner());
            let app_for_checker = app.clone();
            std::thread::spawn(move || completion_checker(state_for_checker, app_for_checker));
            
            true
        }
        Err(_) => false,
    }
}

#[tauri::command]
fn stop_monitor(state: tauri::State<SharedState>) -> bool {
    let mut s = state.lock().unwrap();
    if !s.is_running { return false; }
    
    s.watcher = None;
    s.is_running = false;
    s.start_time.clear();
    s.dir_count = 0;
    s.folder_trackers.clear();
    true
}

#[tauri::command]
fn sign_batch(state: tauri::State<SharedState>, batch_id: String) -> bool {
    let mut s = state.lock().unwrap();
    for batch in &mut s.batches {
        if batch.id == batch_id && batch.status == BatchStatus::Completed {
            batch.status = BatchStatus::Signed;
            batch.signed_at = Some(Local::now().format("%Y-%m-%d %H:%M:%S").to_string());
            save_history(&s.batches);
            return true;
        }
    }
    false
}

#[tauri::command]
fn sign_all_batches(state: tauri::State<SharedState>) {
    let mut s = state.lock().unwrap();
    let now = Local::now().format("%Y-%m-%d %H:%M:%S").to_string();
    for batch in &mut s.batches {
        if batch.status == BatchStatus::Completed {
            batch.status = BatchStatus::Signed;
            batch.signed_at = Some(now.clone());
        }
    }
    save_history(&s.batches);
}

#[tauri::command]
fn clear_batches(state: tauri::State<SharedState>) {
    let mut s = state.lock().unwrap();
    s.batches.retain(|b| b.status != BatchStatus::Signed);
    save_history(&s.batches);
}

#[tauri::command]
fn clear_all_batches(state: tauri::State<SharedState>) {
    let mut s = state.lock().unwrap();
    s.batches.clear();
    save_history(&s.batches);
}

fn handle_file_event(event: Event, state: &SharedState, config: &AppConfig, app: &AppHandle) {
    match event.kind {
        EventKind::Create(_) | EventKind::Modify(_) => {},
        _ => return,
    }
    
    for path in event.paths {
        if path.is_dir() { continue; }
        
        let ext = match path.extension() {
            Some(e) => format!(".{}", e.to_string_lossy().to_lowercase()),
            None => continue,
        };
        
        if !config.file_types.is_empty() && !config.file_types.contains(&ext) {
            continue;
        }
        
        let path_str = path.to_string_lossy().to_string();
        if config.ignore_folders.iter().any(|ig| path_str.contains(&format!("/{}/", ig)) || path_str.contains(&format!("\\{}\\", ig))) {
            continue;
        }
        
        let folder = match path.parent() {
            Some(p) => p.to_string_lossy().to_string(),
            None => continue,
        };
        
        let file_name = path.file_name().map(|n| n.to_string_lossy().to_string()).unwrap_or_default();
        
        let mut s = state.lock().unwrap();
        let now = Instant::now();
        let now_str = Local::now().format("%Y-%m-%d %H:%M:%S").to_string();
        
        let need_notify_start;
        let batch_id_for_update;
        
        if let Some(tracker) = s.folder_trackers.get_mut(&folder) {
            if !tracker.files.contains(&file_name) {
                tracker.files.push(file_name.clone());
            }
            tracker.last_activity = now;
            need_notify_start = !tracker.notified_start;
            if need_notify_start { tracker.notified_start = true; }
            batch_id_for_update = tracker.batch_id.clone();
        } else {
            let batch_id = Uuid::new_v4().to_string();
            let tracker = FolderTracker {
                files: vec![file_name.clone()],
                last_activity: now,
                started_at: now_str.clone(),
                notified_start: true,
                batch_id: batch_id.clone(),
            };
            s.folder_trackers.insert(folder.clone(), tracker);
            
            let batch = Batch {
                id: batch_id.clone(),
                folder: folder.clone(),
                files: vec![file_name.clone()],
                status: BatchStatus::Uploading,
                started_at: now_str,
                completed_at: None,
                signed_at: None,
                file_count: 1,
            };
            s.batches.insert(0, batch);
            if s.batches.len() > 100 { s.batches.truncate(100); }
            
            need_notify_start = true;
            batch_id_for_update = batch_id;
        }
        
        // Update batch files
        let files_clone: Vec<String>;
        if let Some(tracker) = s.folder_trackers.get(&folder) {
            files_clone = tracker.files.clone();
        } else {
            continue;
        }
        
        for batch in &mut s.batches {
            if batch.id == batch_id_for_update {
                batch.files = files_clone.clone();
                batch.file_count = files_clone.len();
                break;
            }
        }
        
        drop(s);
        
        if need_notify_start {
            app.emit("upload-started", &folder).ok();
        }
    }
}

fn completion_checker(state: SharedState, app: AppHandle) {
    let timeout = Duration::from_secs(UPLOAD_COMPLETE_TIMEOUT);
    
    loop {
        std::thread::sleep(Duration::from_secs(1));
        
        let mut s = state.lock().unwrap();
        if !s.is_running { break; }
        
        let now = Instant::now();
        let completed_folders: Vec<String> = s.folder_trackers.iter()
            .filter(|(_, t)| now.duration_since(t.last_activity) >= timeout)
            .map(|(f, _)| f.clone())
            .collect();
        
        for folder in completed_folders {
            if let Some(tracker) = s.folder_trackers.remove(&folder) {
                let completed_time = Local::now().format("%Y-%m-%d %H:%M:%S").to_string();
                for batch in &mut s.batches {
                    if batch.id == tracker.batch_id {
                        batch.status = BatchStatus::Completed;
                        batch.completed_at = Some(completed_time.clone());
                        batch.files = tracker.files.clone();
                        batch.file_count = tracker.files.len();
                        break;
                    }
                }
                
                if s.config.save_history {
                    save_history(&s.batches);
                }
                
                let folder_clone = folder.clone();
                drop(s);
                app.emit("upload-completed", &folder_clone).ok();
                s = state.lock().unwrap();
            }
        }
    }
}

#[cfg_attr(mobile, tauri::mobile_entry_point)]
pub fn run() {
    let config = load_config();
    let batches = load_history();
    
    let state = Arc::new(Mutex::new(GlobalState {
        config,
        batches,
        ..Default::default()
    }));
    
    tauri::Builder::default()
        .plugin(tauri_plugin_opener::init())
        .plugin(tauri_plugin_dialog::init())
        .plugin(tauri_plugin_notification::init())
        .manage(state)
        .invoke_handler(tauri::generate_handler![
            get_config, set_config, get_state,
            start_monitor, stop_monitor,
            sign_batch, sign_all_batches,
            clear_batches, clear_all_batches,
        ])
        .run(tauri::generate_context!())
        .expect("error while running tauri application");
}
