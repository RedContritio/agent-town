import { useEffect, useState, useMemo, useRef } from 'react';
import { Canvas, useFrame } from '@react-three/fiber';
import { OrbitControls, Grid, Text } from '@react-three/drei';
import { worldApi } from '@/api/client';
import type { BlockView, BuildingView, Agent } from '@/types';
import * as THREE from 'three';
import { 
  uiColors, 
  worldColors, 
  buildingColors, 
  agentColors,
  glassStyle 
} from '@/styles/theme';

// Agent component
function AgentMarker({ agent, index }: { agent: Agent; index: number }) {
  const ref = useRef<THREE.Group>(null);
  const color = agentColors[index % agentColors.length];
  
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
      <Text 
        position={[0, 0.8, 0]} 
        fontSize={0.3} 
        color={uiColors.textPrimary}
        anchorX="center" 
        anchorY="middle"
        outlineWidth={0.02}
        outlineColor="black"
      >
        {agent.name}
      </Text>
    </group>
  );
}

// Building with color based on type
function Building({ building }: { building: BuildingView }) {
  const color = buildingColors[building.type] || '#a08060';
  const w = building.width;
  const h = building.height;
  const d = building.depth;
  
  return (
    <group position={[building.anchor.x, 0, building.anchor.y]}>
      {/* Main building */}
      <mesh position={[0, h * 0.5, 0]} castShadow receiveShadow>
        <boxGeometry args={[w, h, d]} />
        <meshStandardMaterial color={color} />
      </mesh>
      
      {/* Pyramid roof */}
      <mesh position={[0, h + h * 0.3, 0]} rotation={[0, Math.PI / 4, 0]} castShadow>
        <coneGeometry args={[Math.max(w, d) * 0.75, h * 0.6, 4]} />
        <meshStandardMaterial color={color} />
      </mesh>
      
      {/* Door */}
      <mesh position={[0, 0.4, d * 0.5 + 0.01]}>
        <boxGeometry args={[0.6, 0.8, 0.05]} />
        <meshStandardMaterial color="#4a3728" />
      </mesh>
      
      <Text 
        position={[0, h + 1, 0]} 
        fontSize={0.35} 
        color={uiColors.textPrimary}
        anchorX="center" 
        anchorY="middle"
        outlineWidth={0.02}
        outlineColor="black"
      >
        {building.name}
      </Text>
    </group>
  );
}

// Terrain color mapping
const terrainColorMap: Record<string, string> = {
  grass: worldColors.grass,
  road: worldColors.road,
  water: worldColors.water,
  farmland: worldColors.farmland,
  sand: worldColors.sand,
  foundation: worldColors.foundation,
  hill: worldColors.hill,
};

// Terrain with height and type-based coloring
function Terrain({ blocks }: { blocks: BlockView[] }) {
  // Group blocks by terrain type
  const blocksByType = useMemo(() => {
    const groups: Record<string, BlockView[]> = {};
    blocks.forEach(block => {
      const type = block.terrainType;
      if (!groups[type]) groups[type] = [];
      groups[type].push(block);
    });
    return groups;
  }, [blocks]);

  return (
    <>
      {Object.entries(blocksByType).map(([type, typeBlocks]) => {
        const color = terrainColorMap[type] || worldColors.grass;
        
        return (
          <InstancedBlocks 
            key={type} 
            blocks={typeBlocks} 
            color={color}
          />
        );
      })}
    </>
  );
}

// Instanced mesh for performance
function InstancedBlocks({ blocks, color }: { blocks: BlockView[]; color: string }) {
  const meshRef = useRef<THREE.InstancedMesh>(null);
  
  useEffect(() => {
    if (!meshRef.current) return;
    
    const matrix = new THREE.Matrix4();
    blocks.forEach((block, i) => {
      // Position based on block coordinates and height
      matrix.setPosition(block.position.x, block.position.z + 0.5, block.position.y);
      meshRef.current!.setMatrixAt(i, matrix);
    });
    meshRef.current.instanceMatrix.needsUpdate = true;
  }, [blocks]);
  
  return (
    <instancedMesh 
      ref={meshRef} 
      args={[undefined, undefined, blocks.length]} 
      castShadow 
      receiveShadow
    >
      <boxGeometry args={[1, 1, 1]} />
      <meshStandardMaterial color={color} />
    </instancedMesh>
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

// Loading screen with HUD style
function LoadingScreen() {
  return (
    <div style={{
      width: '100vw',
      height: '100vh',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      background: uiColors.background,
      color: uiColors.textPrimary,
    }}>
      <h1 style={{ 
        color: uiColors.primary,
        marginBottom: '16px',
        fontSize: '32px',
      }}>Agent Town</h1>
      <div style={{
        width: '200px',
        height: '4px',
        background: uiColors.border,
        borderRadius: '2px',
        overflow: 'hidden',
      }}>
        <div style={{
          width: '60%',
          height: '100%',
          background: uiColors.primary,
          animation: 'loading 1.5s ease-in-out infinite',
        }} />
      </div>
      <p style={{ 
        marginTop: '16px',
        color: uiColors.textSecondary,
      }}>Loading world...</p>
      <style>{`
        @keyframes loading {
          0% { transform: translateX(-100%); }
          50% { transform: translateX(0); }
          100% { transform: translateX(200%); }
        }
      `}</style>
    </div>
  );
}

// Error screen with HUD style
function ErrorScreen({ error, onRetry }: { error: string; onRetry: () => void }) {
  return (
    <div style={{
      width: '100vw',
      height: '100vh',
      display: 'flex',
      flexDirection: 'column',
      alignItems: 'center',
      justifyContent: 'center',
      background: uiColors.background,
      color: uiColors.textPrimary,
      padding: '20px',
    }}>
      <h2 style={{ color: uiColors.danger, marginBottom: '16px' }}>Connection Error</h2>
      <p style={{ color: uiColors.textSecondary, marginBottom: '24px' }}>{error}</p>
      <button 
        onClick={onRetry}
        style={{
          background: 'rgba(79, 195, 247, 0.1)',
          border: `1px solid ${uiColors.primary}`,
          color: uiColors.primary,
          padding: '10px 20px',
          borderRadius: '8px',
          cursor: 'pointer',
          fontSize: '14px',
          transition: 'all 0.25s ease',
        }}
        onMouseEnter={(e) => {
          e.currentTarget.style.background = uiColors.primary;
          e.currentTarget.style.color = uiColors.backgroundDark;
        }}
        onMouseLeave={(e) => {
          e.currentTarget.style.background = 'rgba(79, 195, 247, 0.1)';
          e.currentTarget.style.color = uiColors.primary;
        }}
      >
        Retry
      </button>
    </div>
  );
}

// HUD Panel component
function HudPanel({ title, children, style }: { title: string; children: React.ReactNode; style?: React.CSSProperties }) {
  return (
    <div style={{
      position: 'absolute',
      ...glassStyle,
      color: uiColors.textPrimary,
      padding: '20px',
      minWidth: '200px',
      ...style,
    }}>
      <h2 style={{ 
        margin: '0 0 12px 0', 
        fontSize: '18px',
        color: uiColors.primary,
        borderBottom: `1px solid ${uiColors.borderAccent}`,
        paddingBottom: '8px',
      }}>{title}</h2>
      {children}
    </div>
  );
}

// Stat item component
function StatItem({ label, value }: { label: string; value: string | number }) {
  return (
    <div style={{
      display: 'flex',
      justifyContent: 'space-between',
      alignItems: 'center',
      padding: '4px 0',
      fontSize: '14px',
    }}>
      <span style={{ color: uiColors.textSecondary }}>{label}</span>
      <span style={{ 
        color: uiColors.primary,
        fontFamily: 'Consolas, Monaco, monospace',
        fontWeight: 600,
      }}>{value}</span>
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
      const response = await worldApi.getMap(0, 0, 20);
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
    <div style={{ width: '100vw', height: '100vh', position: 'relative', background: uiColors.background }}>
      {/* HUD Panel - Top Left */}
      <HudPanel 
        title="Agent Town" 
        style={{ top: '20px', left: '20px' }}
      >
        <StatItem label="Agents" value={agents.length} />
        <StatItem label="Buildings" value={buildings.length} />
        <StatItem label="Blocks" value={blocks.length} />
        <div style={{ marginTop: '12px', fontSize: '12px', color: uiColors.textMuted }}>
          World Seed: 123456789
        </div>
      </HudPanel>

      {/* Legend Panel - Top Right */}
      <HudPanel 
        title="Buildings" 
        style={{ top: '20px', right: '20px', minWidth: '150px' }}
      >
        {Object.entries(buildingColors).map(([type, color]) => (
          <div key={type} style={{
            display: 'flex',
            alignItems: 'center',
            gap: '8px',
            padding: '4px 0',
            fontSize: '13px',
          }}>
            <div style={{
              width: '12px',
              height: '12px',
              background: color,
              borderRadius: '2px',
            }} />
            <span style={{ color: uiColors.textSecondary, textTransform: 'capitalize' }}>
              {type.replace('_', ' ')}
            </span>
          </div>
        ))}
      </HudPanel>

      <Canvas camera={{ position: [30, 30, 30], fov: 45 }} shadows>
        <ambientLight intensity={0.5} />
        <directionalLight 
          position={[15, 25, 10]} 
          intensity={1} 
          castShadow 
          shadow-mapSize={[1024, 1024]}
          shadow-camera-far={100}
          shadow-camera-left={-50}
          shadow-camera-right={50}
          shadow-camera-top={50}
          shadow-camera-bottom={-50}
        />
        <Grid 
          args={[100, 100]} 
          cellSize={1} 
          cellThickness={0.3} 
          cellColor={uiColors.borderLight}
          sectionSize={5} 
          sectionThickness={0.8} 
          sectionColor={uiColors.textMuted}
          infiniteGrid 
        />
        <Ground />
        <Terrain blocks={blocks} />
        {buildings.map(b => <Building key={b.id} building={b} />)}
        {agents.map((a, i) => <AgentMarker key={a.id} agent={a} index={i} />)}
        <OrbitControls 
          minDistance={10} 
          maxDistance={100} 
          maxPolarAngle={Math.PI / 2.2}
        />
      </Canvas>
    </div>
  );
}

export default WorldScene;
