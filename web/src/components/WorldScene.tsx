import { Canvas } from '@react-three/fiber';
import { OrbitControls, Grid } from '@react-three/drei';
import { useWorldStore } from '@/stores';

function WorldScene() {
  const { showGrid } = useWorldStore();

  return (
    <div style={{ width: '100vw', height: '100vh' }}>
      <Canvas camera={{ position: [10, 10, 10], fov: 50 }}>
        <ambientLight intensity={0.5} />
        <pointLight position={[10, 10, 10]} />
        
        {showGrid && (
          <Grid
            args={[100, 100]}
            cellSize={1}
            cellThickness={0.5}
            cellColor="#6f6f6f"
            sectionSize={10}
            sectionThickness={1}
            sectionColor="#9d4b4b"
            fadeDistance={50}
            infiniteGrid
          />
        )}
        
        {/* Placeholder for terrain */}
        <mesh position={[0, -0.5, 0]}>
          <boxGeometry args={[10, 1, 10]} />
          <meshStandardMaterial color="#4a6741" />
        </mesh>
        
        {/* Placeholder for buildings */}
        <mesh position={[2, 1, 2]}>
          <boxGeometry args={[2, 2, 2]} />
          <meshStandardMaterial color="#8b7355" />
        </mesh>
        
        {/* Placeholder for agents */}
        <mesh position={[-2, 0.5, -2]}>
          <sphereGeometry args={[0.5, 16, 16]} />
          <meshStandardMaterial color="#e74c3c" />
        </mesh>
        
        <OrbitControls />
      </Canvas>
    </div>
  );
}

export default WorldScene;
