use log::{Level, LevelFilter, Metadata, Record};

use crate::config::LoggerConfig;
use crate::errors::Error;
use serde_json::json;
use time::OffsetDateTime;

pub struct Logger {
    level: Level,
}

impl log::Log for Logger {
    fn enabled(&self, metadata: &Metadata) -> bool {
        metadata.level() <= self.level
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
    pub fn init(config: &LoggerConfig) -> Result<(), Error> {
        let logger = Logger {
            level: config.level,
        };

        log::set_boxed_logger(Box::new(logger))
            .map(|()| log::set_max_level(LevelFilter::Trace))
            .map_err(|err| Error::new(err.to_string()))
    }
}
