SELECT check_status
FROM checks
WHERE uri = ?
ORDER BY ts DESC;