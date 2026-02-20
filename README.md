# GKE Agent

The **GKE Agent** is a specialized, AI-powered CLI tool designed to generate production-ready Kubernetes manifests for Google Kubernetes Engine (GKE). It utilizes the **Gemini 3 Flash** model to understand infrastructure requirements and output valid YAML configurations.

Built with the [Google GenAI ADK for Go](https://github.com/googleapis/genai-toolbox), the agent employs a **multi-agent architecture**:
*   **Root Orchestrator**: Analyzes user requests and routes them to the appropriate specialist.
*   **Specialist Sub-Agents**: Domain-specific agents (e.g., for `ComputeClass` resources) that enforce strict schema compliance and best practices.

## Features

*   **Gemini 3 Powered**: Leverages the latest Gemini 3 Flash Preview model for high-speed, accurate generation.
*   **Vertex AI & ADC Support**: Seamlessly integrates with Google Cloud Vertex AI using Application Default Credentials (ADC).
*   **API Key Fallback**: Supports standard Gemini API keys for non-Vertex environments.
*   **Extensible Architecture**: Easily add new resource specialists by dropping in template files.
*   **GKE Optimization**: Specifically tuned for GKE custom resources (CRDs) like `ComputeClass`.

## Prerequisites

*   Go 1.23 or higher
*   Google Cloud Project (for Vertex AI) OR a Gemini API Key.

## Configuration

The agent is configured via environment variables.

### Option 1: Vertex AI (Recommended)
Uses your local gcloud credentials (ADC).

```bash
export PROJECT_ID="your-google-cloud-project-id"
export LOCATION="us-central1" # or your preferred Vertex AI region
gcloud auth application-default login
```

### Option 2: Gemini API Key
Fallback if Vertex AI variables are not set.

```bash
export GEMINI_API_KEY="your-api-key"
# OR
export GOOGLE_API_KEY="your-api-key"
```

## Build and Run

1.  **Clone the repository** (if applicable) and navigate to the directory:
    ```bash
    cd gke-agent
    ```

2.  **Build the binary**:
    ```bash
    go build -o gke-agent .
    ```

3.  **Run the agent**:
    ```bash
    ./gke-agent "Create a high-performance ComputeClass for a database workload"
    ```

    **Example Output:**
    ```yaml
    [Source: ComputeClass Agent]
    apiVersion: cloud.google.com/v1
    kind: ComputeClass
    metadata:
      name: high-perf-db
    spec:
      priorities:
      - machineType: c3-highmem-8
        storage:
          bootDiskType: hyperdisk-balanced
          bootDiskSizeGb: 200
    ```

## Extending the Agent (Adding Sub-Agents)

The GKE Agent is designed to be easily extensible. To add support for a new Kubernetes resource:

1.  Create a new file in the `templates/sub-agents/` directory.
2.  Name the file `<ResourceName>.tmpl` (e.g., `VerticalPodAutoscaler.tmpl`). The filename becomes the agent's name.
3.  Write the system instructions for the agent in the file.
    *   **Tip**: Be specific! Include the API version, Kind, critical fields, and examples.

**Example `templates/sub-agents/MyResource.tmpl`:**
```text
You are a specialist for MyResource.
OBJECTIVE: Generate valid YAML for MyResource (api/v1).
...
```

The Root Orchestrator will automatically detect the new file on the next run and delegate relevant requests to it.

## Project Structure

*   `main.go`: Entry point. Handles configuration, model initialization, and dynamic loading of sub-agents.
*   `templates/root.tmpl`: System instructions for the main Orchestrator agent.
*   `templates/sub-agents/`: Directory containing instructions for specialized sub-agents.
    *   `ComputeClass.tmpl`: Specialist for GKE ComputeClass resources.
