# Validation Guide

This guide outlines the steps to validate the test deployment.

## 1. Clone Repo

Clone the repository to your local machine:

```bash
git clone <repository-url>
```

## 2. Start Services

Navigate to the repository's root directory and start the services using Docker Compose:

```bash
docker-compose up
```

## 3. Import Sample Data

Import the anonymized AWS and Azure cost data. Detailed instructions will be provided with the sample data files.

## 4. Run Observer

Let the Observer scan the imported data. This process is automatic and may take some time to complete.

## 5. Review AI Recommendations

Once the Observer has finished scanning, review the AI-generated recommendations in the dashboard.
