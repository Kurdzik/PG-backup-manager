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

---

#### Getting Started

**Docker**

1. Clone repo
2. Copy `.env.example` to `.env` and adjust the settings as needed.
3. Run the following command in a project directory:
   ```docker compose up --build -d```
4. This will spin up 
   - Frontend (on port 3000)
   - Backend (on port 8080)
   - Postgres (on port 5432)
   - Minio object storage (on port 9001 - with API port 9002)