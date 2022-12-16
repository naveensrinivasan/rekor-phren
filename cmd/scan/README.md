# Schedule scans for rekor entries

This binary checks the last rekor entry updated in the database and checks it against the rekor server.
If there is a new entry, it will chunk them into 50,000 entries that can be processed by the phren and schedule in k8s.

If there are already running phren jobs, it will not schedule any new jobs.

