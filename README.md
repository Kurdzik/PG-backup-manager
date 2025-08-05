### Postgres Backup Manager
#### Overview
A simple, lightweight application for managing Postgres database backups. You can create backups, store them locally or in an S3-compatible bucket, restore backups, set up periodic backup schedules, and perform other related tasks.

1. Start by connecting to the Postgres database you want to back up.  
   ![Connect to Database](imgs/img1.png)

2. (Optional) Add an S3-compatible destination, such as Minio, for storing backups.  
   ![Add S3 Destination](imgs/img2.png)

3. (Optional) Set a backup schedule. Scheduled jobs currently support backing up databases to an S3 bucket.  
   ![Set Backup Schedule](imgs/img3.png)

4. You can also create backups manually, restore from backups, or view a list of existing backups.  
   ![Manage Backups](imgs/img4.png)