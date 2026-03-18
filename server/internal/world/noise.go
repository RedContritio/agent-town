package world

import (
	"math"
	"math/rand"
)

// SimplexNoise 标准 Simplex Noise 实现
type SimplexNoise struct {
	perm []int
}

// NewSimplexNoise 创建噪声生成器
func NewSimplexNoise(seed int64) *SimplexNoise {
	sn := &SimplexNoise{
		perm: make([]int, 512),
	}
	
	// 使用种子生成排列表
	r := rand.New(rand.NewSource(seed))
	p := make([]int, 256)
	for i := 0; i < 256; i++ {
		p[i] = i
	}
	
	// Fisher-Yates shuffle
	for i := 255; i > 0; i-- {
		j := r.Intn(i + 1)
		p[i], p[j] = p[j], p[i]
	}
	
	// 复制排列表以简化索引
	for i := 0; i < 512; i++ {
		sn.perm[i] = p[i&255]
	}
	
	return sn
}

// 梯度向量表
var grad3 = [][3]float64{
	{1, 1, 0}, {-1, 1, 0}, {1, -1, 0}, {-1, -1, 0},
	{1, 0, 1}, {-1, 0, 1}, {1, 0, -1}, {-1, 0, -1},
	{0, 1, 1}, {0, -1, 1}, {0, 1, -1}, {0, -1, -1},
}

// Noise2D 生成 2D 噪声值 [-1, 1]
func (sn *SimplexNoise) Noise2D(x, y float64) float64 {
	// Skewing/Unskewing 因子
	F2 := 0.5 * (math.Sqrt(3.0) - 1.0)
	G2 := (3.0 - math.Sqrt(3.0)) / 6.0
	
	// Skew 输入空间以确定简单形
	s := (x + y) * F2
	i := int(math.Floor(x + s))
	j := int(math.Floor(y + s))
	t := float64(i+j) * G2
	X0 := float64(i) - t
	Y0 := float64(j) - t
	x0 := x - X0
	y0 := y - Y0
	
	// 确定第二个顶点
	var i1, j1 int
	if x0 > y0 {
		i1 = 1
		j1 = 0
	} else {
		i1 = 0
		j1 = 1
	}
	
	// 其他顶点
	x1 := x0 - float64(i1) + G2
	y1 := y0 - float64(j1) + G2
	x2 := x0 - 1.0 + 2.0*G2
	y2 := y0 - 1.0 + 2.0*G2
	
	// 排列索引
	ii := i & 255
	jj := j & 255
	
	// 计算贡献
	n0, n1, n2 := 0.0, 0.0, 0.0
	
	// 第一个顶点
	t0 := 0.5 - x0*x0 - y0*y0
	if t0 >= 0 {
		t0 *= t0
		gi0 := sn.perm[ii+sn.perm[jj]] % 12
		n0 = t0 * t0 * dot2(grad3[gi0], x0, y0)
	}
	
	// 第二个顶点
	t1 := 0.5 - x1*x1 - y1*y1
	if t1 >= 0 {
		t1 *= t1
		gi1 := sn.perm[ii+i1+sn.perm[jj+j1]] % 12
		n1 = t1 * t1 * dot2(grad3[gi1], x1, y1)
	}
	
	// 第三个顶点
	t2 := 0.5 - x2*x2 - y2*y2
	if t2 >= 0 {
		t2 *= t2
		gi2 := sn.perm[ii+1+sn.perm[jj+1]] % 12
		n2 = t2 * t2 * dot2(grad3[gi2], x2, y2)
	}
	
	// 返回噪声值 [-1, 1]
	return 70.0 * (n0 + n1 + n2)
}

func dot2(g [3]float64, x, y float64) float64 {
	return g[0]*x + g[1]*y
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
