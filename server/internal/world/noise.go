package world

import (
	"math"
)

// SimplexNoise 简化的 Simplex Noise 实现
type SimplexNoise struct {
	seed int64
}

// NewSimplexNoise 创建噪声生成器
func NewSimplexNoise(seed int64) *SimplexNoise {
	return &SimplexNoise{seed: seed}
}

// Noise2D 生成 2D 噪声值 [-1, 1]
func (sn *SimplexNoise) Noise2D(x, y float64) float64 {
	// 基于种子和坐标的哈希噪声
	n := sn.seed + int64(math.Floor(x*374761393)) + int64(math.Floor(y*668265263))
	n = (n ^ (n >> 13)) * 1274126177
	val := float64(n)/float64(^uint(0)>>1) - 1.0
	
	// 平滑插值
	fx := x - math.Floor(x)
	fy := y - math.Floor(y)
	
	// 使用 smoothstep 插值
	sx := fx * fx * (3 - 2*fx)
	sy := fy * fy * (3 - 2*fy)
	
	return val * (1-sx) * (1-sy)
}

// FractalNoise 多层叠加噪声（分形噪声）
// octaves: 层数, persistence: 每层的振幅衰减, lacunarity: 每层的频率倍增
func (sn *SimplexNoise) FractalNoise(x, y float64, octaves int, persistence, lacunarity float64) float64 {
	value := 0.0
	amplitude := 1.0
	frequency := 1.0
	maxValue := 0.0
	
	for i := 0; i < octaves; i++ {
		value += sn.Noise2D(x*frequency, y*frequency) * amplitude
		maxValue += amplitude
		amplitude *= persistence
		frequency *= lacunarity
	}
	
	// 归一化到 [-1, 1]
	if maxValue > 0 {
		value /= maxValue
	}
	
	return value
}

// Noise2DWithSeed 使用指定种子的 2D 噪声（静态方法）
func Noise2DWithSeed(seed int64, x, y float64) float64 {
	sn := NewSimplexNoise(seed)
	return sn.Noise2D(x, y)
}

// FractalNoiseWithSeed 使用指定种子的分形噪声（静态方法）
func FractalNoiseWithSeed(seed int64, x, y float64, octaves int, persistence, lacunarity float64) float64 {
	sn := NewSimplexNoise(seed)
	return sn.FractalNoise(x, y, octaves, persistence, lacunarity)
}
