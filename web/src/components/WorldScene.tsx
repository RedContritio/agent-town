import { useEffect, useState, useMemo, useRef } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { OrbitControls, Grid, Text } from '@react-three/drei';
import { worldApi } from '@/api/client';
import type { BlockView, BuildingView, Agent } from '@/types';
import * as THREE from 'three';

// Agent component
function AgentMarker({ agent }: { agent: Agent }) {
  const ref = useRef<THREE.Group>(null);
  const colors: Record<string, string> = {
    'agent-001': '#ff6b6b',
    'agent-002': '#4ecdc4',
  };
  const color = colors[agent.id] || '#ffe66d';
  
  useFrame((state) => {
    if (ref.current) {
      ref.current.position.y = 0.5 + Math.sin(state.clock.elapsedTime * 3) * 0.05;
    }
  });
  
  return (
    <group ref={ref} position={[agent.position.x, 0.5, agent.position.y]}>
      <mesh castShadow>
        <capsuleGeometry args={[0.25, 0.5, 4, 8]} />
        <meshStandardMaterial color={color} />
      </mesh>
      <Text position={[0, 0.8, 0]} fontSize={0.3} color="white" anchorX="center" anchorY="middle" outlineWidth={0.02} outlineColor="black">
        {agent.name}
      </Text>
    </group>
  );
}

// Building with aligned pyramid roof
function Building({ building }: { building: BuildingView }) {
  const isGovernment = building.ownerId === 'government';
  const wallColor = isGovernment ? '#8b7355' : '#a08060';
  const roofColor = isGovernment ? '#cd853f' : '#d2691e';
  
  const w = building.width;
  const h = building.height;
  const d = building.depth;
  
  return (
    <group position={[building.anchor.x, 0, building.anchor.y]}>
      {/* Main building */}
      <mesh position={[0, h * 0.5, 0]} castShadow receiveShadow>
        <boxGeometry args={[w, h, d]} />
        <meshStandardMaterial color={wallColor} />
      </mesh>
      
      {/* Pyramid roof - tetrahedron with 4 sides, rotated 45° to align edges with box */}
      <mesh position={[0, h + h * 0.3, 0]} rotation={[0, Math.PI / 4, 0]} castShadow>
        <coneGeometry args={[Math.max(w, d) * 0.75, h * 0.6, 4]} />
        <meshStandardMaterial color={roofColor} />
      </mesh>
      
      {/* Door */}
      <mesh position={[0, 0.4, d * 0.5 + 0.01]}>
        <boxGeometry args={[0.6, 0.8, 0.05]} />
        <meshStandardMaterial color="#4a3728" />
      </mesh>
      
      <Text position={[0, h + 1, 0]} fontSize={0.35} color="white" anchorX="center" anchorY="middle" outlineWidth={0.02} outlineColor="black">
        {building.name}
      </Text>
    </group>
  );
}

// Terrain with proper stacking - height 2 = two 1x1x1 blocks
function Terrain({ blocks }: { blocks: BlockView[] }) {
  // Expand blocks - each height level becomes a separate 1x1x1 block
  const expandedBlocks = useMemo(() => {
    const result: { x: number; y: number; z: number; type: string }[] = [];
    
    blocks.forEach(block => {
      const height = block.height;
      // Create stack of 1x1x1 blocks from z=0 to z=height
      for (let h = 0; h <= height; h++) {
        result.push({
          x: block.position.x,
          y: block.position.y,
          z: h, // Each block at its own height level
          type: block.terrainType,
        });
      }
    });
    
    return result;
  }, [blocks]);

  const grassBlocks = useMemo(() => expandedBlocks.filter(b => b.type === 'grass'), [expandedBlocks]);
  const dirtBlocks = useMemo(() => expandedBlocks.filter(b => b.type !== 'grass'), [expandedBlocks]);

  const grassRef = useRef<THREE.InstancedMesh>(null);
  const dirtRef = useRef<THREE.InstancedMesh>(null);
  
  useEffect(() => {
    if (!grassRef.current) return;
    const matrix = new THREE.Matrix4();
    grassBlocks.forEach((block, i) => {
      // Position at integer coordinates, each block sits on top of previous
      matrix.setPosition(block.x, block.z + 0.5, block.y);
      grassRef.current!.setMatrixAt(i, matrix);
    });
    grassRef.current.instanceMatrix.needsUpdate = true;
  }, [grassBlocks]);
  
  useEffect(() => {
    if (!dirtRef.current) return;
    const matrix = new THREE.Matrix4();
    dirtBlocks.forEach((block, i) => {
      matrix.setPosition(block.x, block.z + 0.5, block.y);
      dirtRef.current!.setMatrixAt(i, matrix);
    });
    dirtRef.current.instanceMatrix.needsUpdate = true;
  }, [dirtBlocks]);
  
  return (
    <>
      <instancedMesh ref={grassRef} args={[undefined, undefined, grassBlocks.length]} castShadow receiveShadow>
        <boxGeometry args={[1, 1, 1]} />
        <meshStandardMaterial color="#5a8f4a" />
      </instancedMesh>
      <instancedMesh ref={dirtRef} args={[undefined, undefined, dirtBlocks.length]} castShadow receiveShadow>
        <boxGeometry args={[1, 1, 1]} />
        <meshStandardMaterial color="#8b7355" />
      </instancedMesh>
    </>
  );
}

function Ground() {
  return (
    <mesh position={[0, -0.5, 0]} rotation={[-Math.PI / 2, 0, 0]} receiveShadow>
      <planeGeometry args={[200, 200]} />
      <meshStandardMaterial color="#2c3e50" />
    </mesh>
  );
}

function LoadingScreen() {
  return (
    <div style={{ width: '100vw', height: '100vh', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', background: '#1a1a2e', color: 'white' }}>
      <h1>Agent Town</h1>
      <p>Loading world...</p>
    </div>
  );
}

function ErrorScreen({ error, onRetry }: { error: string; onRetry: () => void }) {
  return (
    <div style={{ width: '100vw', height: '100vh', display: 'flex', flexDirection: 'column', alignItems: 'center', justifyContent: 'center', background: '#1a1a2e', color: 'white', padding: '20px' }}>
      <h2>Connection Error</h2>
      <p>{error}</p>
      <button onClick={onRetry} style={{ padding: '10px 20px', marginTop: '20px' }}>Retry</button>
    </div>
  );
}

function WorldScene() {
  const [blocks, setBlocks] = useState<BlockView[]>([]);
  const [buildings, setBuildings] = useState<BuildingView[]>([]);
  const [agents, setAgents] = useState<Agent[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const fetchData = async () => {
    try {
      setLoading(true);
      const response = await worldApi.getMap(0, 0, 10);
      const data = response.data;
      if (data.blocks) setBlocks(data.blocks);
      if (data.buildings) setBuildings(data.buildings);
      if (data.agents) setAgents(data.agents);
      setError(null);
    } catch (err) {
      setError('Failed to connect to server.');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchData();
    const interval = setInterval(fetchData, 10000);
    return () => clearInterval(interval);
  }, []);

  if (loading && blocks.length === 0) return <LoadingScreen />;
  if (error && blocks.length === 0) return <ErrorScreen error={error} onRetry={fetchData} />;

  return (
    <div style={{ width: '100vw', height: '100vh', position: 'relative' }}>
      <div style={{ position: 'absolute', top: '20px', left: '20px', zIndex: 10, background: 'rgba(0,0,0,0.8)', color: 'white', padding: '20px', borderRadius: '10px' }}>
        <h2>Agent Town</h2>
        <p>Agents: {agents.length}</p>
        <p>Buildings: {buildings.length}</p>
        <p>Blocks: {blocks.length}</p>
      </div>

      <Canvas camera={{ position: [20, 20, 20], fov: 45 }} shadows>
        <ambientLight intensity={0.5} />
        <directionalLight position={[15, 25, 10]} intensity={1} castShadow shadow-mapSize={[1024, 1024]} shadow-camera-far={100} shadow-camera-left={-30} shadow-camera-right={30} shadow-camera-top={30} shadow-camera-bottom={-30} />
        <Grid args={[50, 50]} cellSize={1} cellThickness={0.3} cellColor="#444" sectionSize={5} sectionThickness={0.8} sectionColor="#666" infiniteGrid />
        <Ground />
        <Terrain blocks={blocks} />
        {buildings.map(b => <Building key={b.id} building={b} />)}
        {agents.map(a => <AgentMarker key={a.id} agent={a} />)}
        <OrbitControls minDistance={10} maxDistance={60} maxPolarAngle={Math.PI / 2.2} />
      </Canvas>
    </div>
  );
}

export default WorldScene;
