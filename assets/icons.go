package assets

import (
	"bytes"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"

	"fyne.io/fyne/v2"
)

// AppIcon 应用图标（PNG 格式，兼容系统托盘）
var AppIcon *fyne.StaticResource

// PausedIcon 应用暂停状态图标（PNG 格式）
var PausedIcon *fyne.StaticResource

func init() {
	AppIcon = &fyne.StaticResource{
		StaticName:    "breath-icon.png",
		StaticContent: generateAppIconPNG(),
	}
	PausedIcon = &fyne.StaticResource{
		StaticName:    "breath-paused-icon.png",
		StaticContent: generatePausedIconPNG(),
	}
}

// generateAppIconPNG 生成绿色呼吸图标的 PNG 数据
func generateAppIconPNG() []byte {
	const size = 64
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// 绿色背景圆形
	centerX, centerY := float64(size)/2, float64(size)/2
	radius := float64(size)/2 - 2
	green := color.RGBA{R: 76, G: 175, B: 80, A: 255}       // #4CAF50
	darkGreen := color.RGBA{R: 56, G: 142, B: 60, A: 255}    // #388E3C
	white := color.RGBA{R: 255, G: 255, B: 255, A: 230}      // 白色 opacity 0.9

	// 绘制圆形背景
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - centerX + 0.5
			dy := float64(y) - centerY + 0.5
			dist := math.Sqrt(dx*dx + dy*dy)
			if dist <= radius {
				if dist > radius-2 {
					img.Set(x, y, darkGreen) // 描边
				} else {
					img.Set(x, y, green) // 填充
				}
			}
		}
	}

	// 绘制心形/水滴形状（简化版呼吸符号）
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			fx := (float64(x) - centerX) / 14.0
			fy := (float64(y) - centerY + 2) / 14.0
			// 心形方程简化
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
	// 水滴形状
	if y < -0.5 {
		// 上部圆形
		return x*x+(y+0.5)*(y+0.5) <= 1.0
	}
	// 下部尖角
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
	gray := color.RGBA{R: 158, G: 158, B: 158, A: 255}      // #9E9E9E
	darkGray := color.RGBA{R: 117, G: 117, B: 117, A: 255}   // #757575
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
