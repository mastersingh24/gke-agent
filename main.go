// Copyright 2026 The Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"google.golang.org/adk/agent"
	"google.golang.org/adk/agent/llmagent"
	"google.golang.org/adk/cmd/launcher"
	"google.golang.org/adk/cmd/launcher/full"
	"google.golang.org/adk/model"
	"google.golang.org/adk/model/gemini"
	"google.golang.org/adk/tool"
	"google.golang.org/adk/tool/agenttool"
	"google.golang.org/genai"
)

// Config holds the paths for templates
const (
	rootTemplatePath = "templates/root.tmpl"
	subAgentsDir     = "templates/sub-agents"
	modelName        = "gemini-3-flash-preview"
)

func main() {
	ctx := context.Background()

	// 1. Initialize the Model
	// Check for Vertex AI configuration
	projectID := os.Getenv("GOOGLE_CLOUD_PROJECT")
	location := os.Getenv("GOOGLE_CLOUD_LOCATION")

	var clientConfig *genai.ClientConfig
	if projectID != "" && location != "" {
		clientConfig = &genai.ClientConfig{
			Project:  projectID,
			Location: location,
			Backend:  genai.BackendVertexAI,
		}
	} else {
		// Fallback to API Key
		apiKey := os.Getenv("GEMINI_API_KEY")
		if apiKey == "" {
			apiKey = os.Getenv("GOOGLE_API_KEY")
		}
		clientConfig = &genai.ClientConfig{
			APIKey: apiKey,
		}
	}

	model, err := gemini.NewModel(ctx, modelName, clientConfig)
	if err != nil {
		log.Fatalf("Failed to create model: %v", err)
	}

	// 2. Load Sub-Agents dynamically from the templates directory
	subAgents, err := loadSubAgents(ctx, model, subAgentsDir)
	if err != nil {
		log.Fatalf("Failed to load sub-agents: %v", err)
	}

	// 3. Convert Sub-Agents into Tools for the Root Agent
	// This allows the Root Agent to delegate specific resource generation to specialists.
	var agentTools []tool.Tool
	for _, sa := range subAgents {
		t := agenttool.New(sa, nil)
		agentTools = append(agentTools, t)
	}

	// 4. Load Root Agent Instruction
	rootInstruction, err := os.ReadFile(rootTemplatePath)
	if err != nil {
		log.Fatalf("Failed to read root template: %v", err)
	}

	// 5. Create the Root Orchestrator Agent
	rootAgent, err := llmagent.New(llmagent.Config{
		Name:        "K8sOrchestrator",
		Model:       model,
		Description: "A specialized orchestrator for generating Kubernetes manifests. It delegates to sub-agents for specific resource types.",
		Instruction: string(rootInstruction),
		Tools:       agentTools,
	})
	if err != nil {
		log.Fatalf("Failed to create root agent: %v", err)
	}

	// 6. Launch the Agent
	config := &launcher.Config{
		AgentLoader: agent.NewSingleLoader(rootAgent),
	}
	l := full.NewLauncher()
	if err = l.Execute(ctx, config, os.Args[1:]); err != nil {
		log.Fatalf("Agent execution failed: %v", err)
	}
}

// loadSubAgents reads the templates directory and instantiates specialized agents.
func loadSubAgents(ctx context.Context, model model.LLM, dir string) ([]agent.Agent, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var agents []agent.Agent
	for _, file := range files {
		if file.IsDir() || !strings.HasSuffix(file.Name(), ".tmpl") {
			continue
		}

		path := filepath.Join(dir, file.Name())
		content, err := os.ReadFile(path)
		if err != nil {
			return nil, err
		}

		// The filename (minus extension) acts as the Agent Name/Description for tool calling logic
		agentName := strings.TrimSuffix(file.Name(), ".tmpl")
		log.Printf("Loading sub-agent: %s", agentName)

		sa, err := llmagent.New(llmagent.Config{
			Name:        agentName,
			Model:       model,
			Description: fmt.Sprintf("Specialist agent for generating %s Kubernetes manifests.", agentName),
			Instruction: string(content),
		})
		if err != nil {
			return nil, err
		}
		agents = append(agents, sa)
	}

	return agents, nil
}
