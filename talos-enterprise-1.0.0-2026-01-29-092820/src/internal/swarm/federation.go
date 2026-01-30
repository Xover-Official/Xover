package swarm

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"
)

// Insight represents a learned pattern to be shared
type Insight struct {
	SourceNodeID string    `json:"source_node"`
	ActionType   string    `json:"action_type"`
	ContextHash  string    `json:"context_hash"` // Anonymized feature vector
	Outcome      float64   `json:"outcome"`      // +1 success, -1 fail
	Confidence   float64   `json:"confidence"`
	Timestamp    time.Time `json:"timestamp"`
}

// FederatedBrain connects multiple Talos nodes
type FederatedBrain struct {
	NodeID         string
	Peers          []string
	KnowledgeGraph map[string]float64 // Shared action weights
	mu             sync.RWMutex
}

func NewFederatedBrain(nodeID string) *FederatedBrain {
	return &FederatedBrain{
		NodeID:         nodeID,
		KnowledgeGraph: make(map[string]float64),
	}
}

// BroadcastInsight shares a local learning with the swarm
func (f *FederatedBrain) BroadcastInsight(ctx context.Context, insight Insight) {
	// In production, this uses P2P or a central coordinator (Redis/Kafka)
	// Mocking broadcast
	fmt.Printf("📡 FEDERATION: Node %s sharing insight on %s (Confidence: %.2f)\n",
		f.NodeID, insight.ActionType, insight.Confidence)
}

// AbsorbKnowledge integrates insights from peers
func (f *FederatedBrain) AbsorbKnowledge(insights []Insight) {
	f.mu.Lock()
	defer f.mu.Unlock()

	for _, insight := range insights {
		// Federated Averaging (FedAvg) logic simplified
		currentScore := f.KnowledgeGraph[insight.ActionType]

		// Weighted update based on peer confidence
		newScore := (currentScore + (insight.Outcome * insight.Confidence)) / 2

		f.KnowledgeGraph[insight.ActionType] = newScore
		log.Printf("🧠 SYNAPSE: Updated collective weight for %s: %.4f", insight.ActionType, newScore)
	}
}
