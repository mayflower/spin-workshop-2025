use std::time::{SystemTime, UNIX_EPOCH};

use spin_cron_sdk::{cron_component, Error, Metadata};
use spin_sdk::{
    sqlite::{self, Connection},
    variables,
};

#[cron_component]
fn handle_cron_event(_metadata: Metadata) -> Result<(), Error> {
    let conn = Connection::open_default()
        .map_err(|e| Error::Other(format!("failed to open database: {}", e.to_string())))?;

    let max_age_seconds = variables::get("max_age_seconds")
        .map_err(|e| Error::Other(format!("failed to get max_age_seconds: {}", e.to_string())))?
        .parse::<i64>()
        .map_err(|e| Error::Other(format!("failed to parse max_age_seconds {}", e.to_string())))?;

    let now = SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .map(|x| x.as_secs())
        .map_err(|e| Error::Other(format!("failed to get epoch: {}", e.to_string())))?
        as i64;

    println!("evicting all documents older than {} seconds", max_age_seconds);

    conn.execute(
        "DELETE FROM derived WHERE last_access < ?;",
        [sqlite::Value::Integer(now - max_age_seconds)].as_slice(),
    )
    .map(|_| ())
    .map_err(|e| Error::Other(format!("query failed: {}", e.to_string())))
}
