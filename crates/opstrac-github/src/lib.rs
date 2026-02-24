use serde::{Deserialize, Serialize};

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GitHubIntegration {
    pub owner: String,
    pub repo: String,
    pub token: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct GitHubIncident {
    pub issue_number: u64,
    pub title: String,
    pub body: String,
    pub labels: Vec<String>,
    pub state: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub struct DeploymentEvent {
    pub environment: String,
    pub sha: String,
    pub status: DeploymentStatus,
    pub description: String,
}

#[derive(Debug, Clone, Serialize, Deserialize)]
pub enum DeploymentStatus {
    Pending,
    InProgress,
    Success,
    Failure,
    Error,
}

impl GitHubIntegration {
    pub fn new(owner: String, repo: String, token: String) -> Self {
        Self { owner, repo, token }
    }

    pub async fn create_incident_issue(&self, title: &str, body: &str, labels: Vec<String>) -> Result<u64, Box<dyn std::error::Error + Send + Sync>> {
        let octocrab = octocrab::Octocrab::builder()
            .personal_token(self.token.clone())
            .build()?;
        let issue = octocrab.issues(&self.owner, &self.repo)
            .create(title)
            .body(body)
            .labels(labels)
            .send()
            .await?;
        Ok(issue.number)
    }

    pub async fn close_incident_issue(&self, issue_number: u64) -> Result<(), Box<dyn std::error::Error + Send + Sync>> {
        let octocrab = octocrab::Octocrab::builder()
            .personal_token(self.token.clone())
            .build()?;
        octocrab.issues(&self.owner, &self.repo)
            .update(issue_number)
            .state(octocrab::params::State::Closed)
            .send()
            .await?;
        Ok(())
    }

    pub async fn list_deployments(&self) -> Result<Vec<DeploymentEvent>, Box<dyn std::error::Error + Send + Sync>> {
        tracing::info!("Listing deployments for {}/{}", self.owner, self.repo);
        Ok(vec![])
    }
}
