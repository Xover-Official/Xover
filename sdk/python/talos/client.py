"""
Talos Guardian Python SDK
Official Python client for the Talos autonomous cloud optimization platform.
"""

import requests
from typing import Dict, List, Optional, Any
from dataclasses import dataclass
from datetime import datetime


@dataclass
class TierStatus:
    """Status of an AI tier"""
    tier: int
    name: str
    model: str
    active: bool
    requests_today: int
    avg_latency_ms: float
    success_rate: float
    status: str


@dataclass
class SwarmStatus:
    """AI Swarm status"""
    active_tier: int
    tier_status: List[TierStatus]
    current_action: str
    queue_depth: int


@dataclass
class Resource:
    """Cloud resource"""
    id: str
    type: str
    provider: str
    region: str
    cost_per_month: float
    optimization_score: float
    tags: Dict[str, str]


@dataclass
class ROI:
    """ROI metrics"""
    ratio: float
    total_savings: float
    total_cost: float
    net_profit: float


class TalosClient:
    """Talos API Client"""
    
    def __init__(self, base_url: str = "http://localhost:8080", api_key: Optional[str] = None):
        """
        Initialize Talos client
        
        Args:
            base_url: Talos backend URL
            api_key: Optional API key for authentication
        """
        self.base_url = base_url.rstrip('/')
        self.api_key = api_key
        self.session = requests.Session()
        
        if api_key:
            self.session.headers.update({
                'Authorization': f'Bearer {api_key}'
            })
    
    def health(self) -> Dict[str, Any]:
        """Check system health"""
        response = self.session.get(f'{self.base_url}/health')
        response.raise_for_status()
        return response.json()
    
    def get_swarm_status(self) -> SwarmStatus:
        """Get AI swarm status"""
        response = self.session.get(f'{self.base_url}/api/swarm/live')
        response.raise_for_status()
        data = response.json()
        
        return SwarmStatus(
            active_tier=data['active_tier'],
            tier_status=[TierStatus(**t) for t in data['tier_status']],
            current_action=data['current_action'],
            queue_depth=data['queue_depth']
        )
    
    def run_optimization(self, 
                        type: str = "full",
                        risk_limit: float = 7.0,
                        dry_run: bool = True) -> Dict[str, Any]:
        """
        Run optimization
        
        Args:
            type: Optimization type ('full', 'cost', 'performance')
            risk_limit: Maximum risk score (0-10)
            dry_run: If True, only simulate changes
            
        Returns:
            Optimization results
        """
        payload = {
            'type': type,
            'risk_limit': risk_limit,
            'dry_run': dry_run
        }
        
        response = self.session.post(
            f'{self.base_url}/api/optimize',
            json=payload
        )
        response.raise_for_status()
        return response.json()
    
    def get_resources(self, 
                     provider: Optional[str] = None,
                     resource_type: Optional[str] = None) -> List[Resource]:
        """
        Get cloud resources
        
        Args:
            provider: Filter by provider ('aws', 'gcp', 'azure')
            resource_type: Filter by resource type
            
        Returns:
            List of resources
        """
        params = {}
        if provider:
            params['provider'] = provider
        if resource_type:
            params['type'] = resource_type
        
        response = self.session.get(
            f'{self.base_url}/api/resources',
            params=params
        )
        response.raise_for_status()
        
        return [Resource(**r) for r in response.json()]
    
    def get_roi(self) -> ROI:
        """Get ROI metrics"""
        response = self.session.get(f'{self.base_url}/api/roi')
        response.raise_for_status()
        data = response.json()
        
        return ROI(**data)
    
    def chat(self, message: str) -> str:
        """
        Chat with AI swarm
        
        Args:
            message: Chat message
            
        Returns:
            AI response
        """
        response = self.session.post(
            f'{self.base_url}/api/ai/chat',
            json={'message': message}
        )
        response.raise_for_status()
        return response.json()['response']
    
    def get_recommendations(self) -> List[Dict[str, Any]]:
        """Get optimization recommendations"""
        response = self.session.get(f'{self.base_url}/api/recommendations')
        response.raise_for_status()
        return response.json()


# Example usage
if __name__ == '__main__':
    # Initialize client
    client = TalosClient(
        base_url='http://localhost:8080',
        api_key='your-api-key'  # Optional
    )
    
    # Check health
    health = client.health()
    print(f"System status: {health['status']}")
    
    # Get AI swarm status
    swarm = client.get_swarm_status()
    print(f"Active tier: T{swarm.active_tier}")
    print(f"Current action: {swarm.current_action}")
    
    # Get ROI
    roi = client.get_roi()
    print(f"ROI: {roi.ratio:.1f}x")
    print(f"Savings: ${roi.total_savings:.2f}")
    
    # Run optimization (dry-run)
    result = client.run_optimization(dry_run=True)
    print(f"Found {result['optimizations_found']} optimizations")
    print(f"Est. savings: ${result['estimated_savings']:.2f}")
    
    # Chat with AI
    response = client.chat("Should I optimize my AWS resources?")
    print(f"AI: {response}")
