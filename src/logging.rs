use log::{Level, LevelFilter, Metadata, Record};

use crate::errors::Error;
use serde_json::json;
use time::OffsetDateTime;

pub struct Logger;

impl log::Log for Logger {
    fn enabled(&self, metadata: &Metadata) -> bool {
        metadata.level() <= Level::Info
    }

    fn log(&self, record: &Record) {
        if self.enabled(record.metadata()) {
            eprintln!(
                "{}",
                json!({
                    "time": OffsetDateTime::now_utc().to_string(),
                    "level": record.level(),
                    "message": record.args(),
                })
            )
        }
    }

    fn flush(&self) {}
}

impl Logger {
    // TODO: Have this configured from the logging config
    pub fn init() -> Result<(), Error> {
        log::set_boxed_logger(Box::new(Logger))
            .map(|()| log::set_max_level(LevelFilter::Info))
            .map_err(|err| Error::new(err.to_string()))
    }
}
