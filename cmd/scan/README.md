# Schedule scans for rekor entries

This binary checks the last `rekor` entry updated in the database and checks it against the `rekor` server.
If there is a new entry, it will chunk the entries that need to be updated into `50,000` that can be
processed by the phren and schedule in k8s.

If there are already running phren jobs, it will not schedule any new jobs.

This is scheduled as a k8s cronjob.

The job is partitioned to process in parallel. 

