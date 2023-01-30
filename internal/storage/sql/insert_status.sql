INSERT INTO checks (ts, uri, check_status, err_msg)
VALUES (?, ?, ?, ?)
RETURNING id;