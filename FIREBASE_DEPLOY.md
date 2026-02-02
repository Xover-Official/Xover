# 🔥 Deploying Talos Atlas to Firebase

Talos Atlas is ready for deployment on the Google Cloud / Firebase stack.

## 🚀 Quick Deployment Guide

### Prerequisites

- [Firebase CLI](https://firebase.google.com/docs/cli) installed (`npm install -g firebase-tools`)
- A Google Cloud Project created at [console.firebase.google.com](https://console.firebase.google.com)

### 1. Initialize & Login

```bash
firebase login
firebase use --add  # Select your project
```

### 2. Deployment Architecture

- **Frontend**: The `web/` directory is deployed to **Firebase Hosting** (Global CDN).
- **Backend**: The Go application (`cmd/dashboard`) runs on **Cloud Run** or as a **Cloud Function**.

### 3. Deploying the Frontend

To deploy the static dashboard and login portal:

```bash
firebase deploy --only hosting
```

Your app will be live at `https://<your-project>.web.app`.

### 4. Deploying the Backend (Cloud Run)

For the API (`/api/...`) to work, deploy the Go backend container to Cloud Run:

1. **Build Container**:

   ```bash
   gcloud builds submit --tag gcr.io/PROJECT_ID/talos-dashboard
   ```

2. **Deploy Service**:

   ```bash
   gcloud run deploy talos-api \
     --image gcr.io/PROJECT_ID/talos-dashboard \
     --platform managed \
     --allow-unauthenticated
   ```

3. **Connect Hosting to Backend**:
   Update `firebase.json` to point `/api/**` rewrites to your Cloud Run service instead of a function, or configure Firebase Hosting to rewrite to the Cloud Run service directly.

   *Example `firebase.json` update for Cloud Run integration:*

   ```json
   "rewrites": [ {
     "source": "/api/**",
     "run": {
       "serviceId": "talos-api",
       "region": "us-central1"
     }
   } ]
   ```

## 🛡️ Security Note

This configuration assumes `web/` contains public assets. Ensure no sensitive Go source code or `.env` files are in the `web/` directory (they are currently separate, which is good).
