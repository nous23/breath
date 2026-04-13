package assets

import (
	"bytes"
	"encoding/binary"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
)

// AppIconPNG 应用图标 PNG 数据（用于系统托盘，macOS/Linux）
var AppIconPNG []byte

// PausedIconPNG 暂停状态图标 PNG 数据（macOS/Linux）
var PausedIconPNG []byte

// AppIconICO 应用图标 ICO 数据（用于系统托盘，Windows）
var AppIconICO []byte

// PausedIconICO 暂停状态图标 ICO 数据（Windows）
var PausedIconICO []byte

func init() {
	AppIconPNG = generateAppIconPNG()
	PausedIconPNG = generatePausedIconPNG()
	AppIconICO = pngToICO(AppIconPNG)
	PausedIconICO = pngToICO(PausedIconPNG)
}

// generateAppIconPNG 生成绿色呼吸图标的 PNG 数据
func generateAppIconPNG() []byte {
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// 绿色背景圆形
	centerX, centerY := float64(size)/2, float64(size)/2
	radius := float64(size)/2 - 2
	green := color.RGBA{R: 76, G: 175, B: 80, A: 255}
	darkGreen := color.RGBA{R: 56, G: 142, B: 60, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 230}

	// 绘制圆形背景
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX + 0.5
			dy := float64(y) - centerY + 0.5
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= radius {
				if dist > radius-2 {
					img.Set(x, y, darkGreen)
				} else {
					img.Set(x, y, green)
				}
			}
		}
	}

	// 绘制心形/水滴形状（简化版呼吸符号）
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fx := (float64(x) - centerX) / 14.0
			fy := (float64(y) - centerY + 2) / 14.0
			if isInHeart(fx, fy) {
				img.Set(x, y, white)
			}
		}
	}

	// 绘制中心小绿点
	smallR := 4.0
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX + 0.5
			dy := float64(y) - (centerY - 4) + 0.5
			if math.Sqrt(dx*dx+dy*dy) <= smallR {
				img.Set(x, y, green)
			}
		}
	}

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// isInHeart 判断点是否在心形/水滴区域内
func isInHeart(x, y float64) bool {
	if y < -0.5 {
		return x*x+(y+0.5)*(y+0.5) <= 1.0
	}
	t := (y + 0.5) / 2.0
	if t > 1.0 {
		return false
	}
	halfWidth := 1.0 - t
	return x >= -halfWidth && x <= halfWidth && y <= 1.8
}

// generatePausedIconPNG 生成灰色暂停图标的 PNG 数据
func generatePausedIconPNG() []byte {
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	centerX, centerY := float64(size)/2, float64(size)/2
	radius := float64(size)/2 - 2
	gray := color.RGBA{R: 158, G: 158, B: 158, A: 255}
	darkGray := color.RGBA{R: 117, G: 117, B: 117, A: 255}
	white := color.RGBA{R: 255, G: 255, B: 255, A: 230}

	// 绘制圆形背景
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX + 0.5
			dy := float64(y) - centerY + 0.5
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= radius {
				if dist > radius-2 {
					img.Set(x, y, darkGray)
				} else {
					img.Set(x, y, gray)
				}
			}
		}
	}

	// 绘制两个暂停竖条
	pauseRect1 := image.Rect(22, 20, 30, 44)
	pauseRect2 := image.Rect(34, 20, 42, 44)
	draw.Draw(img, pauseRect1, &image.Uniform{white}, image.Point{}, draw.Over)
	draw.Draw(img, pauseRect2, &image.Uniform{white}, image.Point{}, draw.Over)

	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}

// pngToICO 将 PNG 数据转换为 ICO 格式
// ICO 格式支持内嵌 PNG 数据（Vista+ 格式），无需解码再编码为 BMP
func pngToICO(pngData []byte) []byte {
	// 解码 PNG 获取尺寸信息
	img, err := png.Decode(bytes.NewReader(pngData))
	if err != nil {
		return pngData // 降级返回 PNG
	}
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// ICO 文件头（6 字节）
	// ICO 目录条目（16 字节）
	// PNG 数据
	var buf bytes.Buffer

	// ICONDIR 头
	binary.Write(&buf, binary.LittleEndian, uint16(0))    // 保留，必须为 0
	binary.Write(&buf, binary.LittleEndian, uint16(1))    // 类型：1 = ICO
	binary.Write(&buf, binary.LittleEndian, uint16(1))    // 图像数量：1

	// ICONDIRENTRY 条目
	// 宽度和高度：0 表示 256
	w := uint8(width)
	h := uint8(height)
	if width >= 256 {
		w = 0
	}
	if height >= 256 {
		h = 0
	}
	buf.WriteByte(w)                                                    // 宽度
	buf.WriteByte(h)                                                    // 高度
	buf.WriteByte(0)                                                    // 调色板颜色数（0 = 无调色板）
	buf.WriteByte(0)                                                    // 保留
	binary.Write(&buf, binary.LittleEndian, uint16(1))                 // 色彩平面数
	binary.Write(&buf, binary.LittleEndian, uint16(32))                // 每像素位数
	binary.Write(&buf, binary.LittleEndian, uint32(len(pngData)))      // 图像数据大小
	binary.Write(&buf, binary.LittleEndian, uint32(6+16))              // 图像数据偏移（头6 + 条目16）

	// 写入 PNG 数据
	buf.Write(pngData)

	return buf.Bytes()
}