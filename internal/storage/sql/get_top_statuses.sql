SELECT id, ts, uri, check_status
FROM checks
ORDER BY ts DESC
LIMIT 10
OFFSET ?;