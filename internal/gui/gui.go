//go:build windows

package gui

import (
	"fmt"
	"path/filepath"
	"strings"
	"sync/atomic"
	"time"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
	"github.com/pengyongshi/cutx/internal/chunk"
	"github.com/pengyongshi/cutx/internal/engine"
)

type reporter struct {
	mw      *walk.MainWindow
	pb      *walk.ProgressBar
	lt      *walk.TextEdit
	st      *walk.Label
	total   int64
	done    int64
	running int32
}

func (r *reporter) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	r.mw.Synchronize(func() {
		if r.lt != nil {
			cur := r.lt.Text()
			if cur != "" {
				cur += "\r\n"
			}
			ts := time.Now().Format("15:04:05")
			r.lt.SetText(cur + "[" + ts + "] " + msg)
		}
	})
}

func (r *reporter) ProgressStart(total int64, label string) {
	atomic.StoreInt64(&r.total, total)
	atomic.StoreInt64(&r.done, 0)
	r.mw.Synchronize(func() {
		if r.pb != nil {
			r.pb.SetRange(0, int(total))
			r.pb.SetValue(0)
		}
		if r.st != nil {
			r.st.SetText(label)
		}
	})
}

func (r *reporter) ProgressAdd(n int64) {
	cur := atomic.AddInt64(&r.done, n)
	r.mw.Synchronize(func() {
		if r.pb != nil {
			r.pb.SetValue(int(cur))
		}
	})
}

func (r *reporter) ProgressFinish() {
	r.mw.Synchronize(func() {
		if r.pb != nil {
			r.pb.SetValue(int(atomic.LoadInt64(&r.total)))
		}
	})
}

func (r *reporter) ProgressStop() {}

var _ engine.Reporter = (*reporter)(nil)

// Run starts the Windows GUI.
func Run(version string) {
	var mw *walk.MainWindow
	var pb *walk.ProgressBar
	var logTE *walk.TextEdit
	var statusLbl *walk.Label

	var fileLE *walk.LineEdit
	var opCB *walk.ComboBox
	var sizeCB *walk.ComboBox
	var hashCB *walk.ComboBox
	var outLE *walk.LineEdit
	var modeCB *walk.ComboBox
	var forceCB *walk.CheckBox
	var startBtn *walk.PushButton
	var previewLbl *walk.Label

	rep := &reporter{}

	mw_ := MainWindow{
		AssignTo: &mw,
		Title:    "CutX — 文件切割与合并工具 " + version,
		MinSize:  Size{Width: 600, Height: 580},
		Size:     Size{Width: 640, Height: 640},
		Layout:   VBox{MarginsZero: true, Spacing: 0},
		Children: []Widget{
			// Main content
			Composite{
				Layout:   VBox{MarginsZero: false, Spacing: 12},
				Children: []Widget{
					// Title
					Label{
						Text:  "  CutX — 文件切割与合并工具",
						Font:  Font{Family: "Segoe UI", PointSize: 14, Bold: true},
						MinSize: Size{Height: 36},
					},

					// Operation + File path
					Composite{
						Layout: Grid{Columns: 2, Spacing: 10},
						Children: []Widget{
							Label{Text: "操作:", Font: Font{PointSize: 9}},
							ComboBox{
								AssignTo: &opCB,
								Value:    "切割 Split",
								Model:    []string{"切割 Split", "合并 Merge", "校验 Verify"},
								Font:     Font{PointSize: 9},
								OnCurrentIndexChanged: func() {
									updateUI(opCB, sizeCB, hashCB, outLE, modeCB, forceCB, previewLbl, fileLE)
								},
							},

							Label{Text: "文件路径:", Font: Font{PointSize: 9}},
							Composite{
								Layout: HBox{MarginsZero: true, Spacing: 4},
								Children: []Widget{
									LineEdit{AssignTo: &fileLE, Font: Font{Family: "Consolas", PointSize: 9},
										OnTextChanged: func() {
											updateUI(opCB, sizeCB, hashCB, outLE, modeCB, forceCB, previewLbl, fileLE)
										}},
									PushButton{
										Text: "浏览...",
										Font: Font{PointSize: 9},
										OnClicked: func() {
											pickFile(mw, fileLE, opCB)
											updateUI(opCB, sizeCB, hashCB, outLE, modeCB, forceCB, previewLbl, fileLE)
										},
									},
								},
							},
						},
					},

					// Split params
					Composite{
						Layout: Grid{Columns: 2, Spacing: 10},
						Children: []Widget{
							Label{Text: "切割大小:", Font: Font{PointSize: 9}},
							ComboBox{
								AssignTo: &sizeCB,
								Editable: true,
								Value:    "2G",
								Model:    []string{"2G", "1G", "500M", "100M", "50M", "10M", "1M"},
								Font:     Font{PointSize: 9},
							},

							Label{Text: "校验算法:", Font: Font{PointSize: 9}},
							ComboBox{
								AssignTo: &hashCB,
								Value:    "md5 (快速)",
								Model:    []string{"md5 (快速)", "sha256 (安全)"},
								Font:     Font{PointSize: 9},
							},

							Label{Text: "输出目录:", Font: Font{PointSize: 9}},
							LineEdit{
								AssignTo:    &outLE,
								CueBanner:   "留空 = 自动创建 cutx-<文件名>/ 子目录",
								Font:        Font{Family: "Consolas", PointSize: 9},
							},
						},
					},

					// Merge params
					Composite{
						Layout: Grid{Columns: 2, Spacing: 10},
						Children: []Widget{
							Label{Text: "合并模式:", Font: Font{PointSize: 9}},
							ComboBox{
								AssignTo: &modeCB,
								Value:    "quick (快速，不校验)",
								Model:    []string{"quick (快速，不校验)", "verify (边合并边校验)"},
								Font:     Font{PointSize: 9},
							},
							Label{Text: ""},
							CheckBox{
								AssignTo: &forceCB,
								Text:     "强制覆盖已存在的文件",
								Font:     Font{PointSize: 9},
							},
						},
					},

					// Preview
					Label{
						AssignTo: &previewLbl,
						Text:     "",
						Font:     Font{PointSize: 8},
					},

					// Start button
					PushButton{
						AssignTo: &startBtn,
						Text:     "▶ 开始",
						Font:     Font{PointSize: 11, Bold: true},
						MinSize:  Size{Height: 40},
						OnClicked: func() {
							if atomic.LoadInt32(&rep.running) != 0 {
								walk.MsgBox(mw, "提示", "有操作正在进行中", walk.MsgBoxIconWarning)
								return
							}
							rep.mw = mw
							rep.pb = pb
							rep.lt = logTE
							rep.st = statusLbl
							atomic.StoreInt32(&rep.running, 1)
							startBtn.SetEnabled(false)
							go func() {
								runOperation(rep, opCB, fileLE, sizeCB, hashCB, outLE, modeCB, forceCB, version)
								atomic.StoreInt32(&rep.running, 0)
								mw.Synchronize(func() {
									startBtn.SetEnabled(true)
								})
							}()
						},
					},
				},
			},

			// Progress bar
			Composite{
				Layout:  HBox{MarginsZero: true, Spacing: 10},
				MinSize: Size{Height: 36},
				Children: []Widget{
					ProgressBar{AssignTo: &pb, MinSize: Size{Width: 200}},
					Label{AssignTo: &statusLbl, Text: "就绪", Font: Font{PointSize: 9}},
				},
			},

			// Log area
			GroupBox{
				Title:    "日志",
				Layout:   VBox{MarginsZero: true},
				Children: []Widget{
					TextEdit{
						AssignTo:  &logTE,
						ReadOnly:  true,
						VScroll:   true,
						MinSize:   Size{Height: 120},
						Font:      Font{Family: "Consolas", PointSize: 9},
					},
				},
			},
		},
	}


	if _, err := mw_.Run(); err != nil {
		fmt.Println("GUI error:", err)
	}
}

func updateUI(opCB, sizeCB, hashCB *walk.ComboBox, outLE *walk.LineEdit,
	modeCB *walk.ComboBox, forceCB *walk.CheckBox,
	previewLbl *walk.Label, fileLE *walk.LineEdit) {

	op := opCB.Text()
	isSplit := strings.HasPrefix(op, "切割")
	isMerge := strings.HasPrefix(op, "合并")

	sizeCB.SetEnabled(isSplit)
	hashCB.SetEnabled(isSplit)
	outLE.SetEnabled(isSplit || isMerge)
	modeCB.SetEnabled(isMerge)
	forceCB.SetEnabled(isMerge)

	if isSplit && fileLE != nil {
		fp := fileLE.Text()
		if fp != "" {
			out := outLE.Text()
			if out != "" {
				previewLbl.SetText("输出到: " + out)
			} else {
				base := chunk.BaseName(fp)
				stripped := chunk.StripExtensions(base)
				dir := filepath.Dir(fp)
				previewLbl.SetText(fmt.Sprintf("将创建: %s\\cutx-%s\\", dir, stripped))
			}
		} else {
			previewLbl.SetText("")
		}
	} else {
		previewLbl.SetText("")
	}
}

func pickFile(mw *walk.MainWindow, fileLE *walk.LineEdit, opCB *walk.ComboBox) {
	dlg := &walk.FileDialog{Title: "选择文件"}
	op := opCB.Text()
	if strings.HasPrefix(op, "校验") || strings.HasPrefix(op, "合并") {
		dlg.Filter = "Manifest files (*.manifest.json)|*.manifest.json|JSON files (*.json)|*.json|All files (*.*)|*.*"
	} else {
		dlg.Filter = "All files (*.*)|*.*"
	}
	if ok, _ := dlg.ShowOpen(mw); ok {
		fileLE.SetText(dlg.FilePath)
	}
}

func runOperation(rep *reporter,
	opCB *walk.ComboBox, fileLE *walk.LineEdit,
	sizeCB *walk.ComboBox, hashCB *walk.ComboBox,
	outLE *walk.LineEdit, modeCB *walk.ComboBox, forceCB *walk.CheckBox,
	version string) {

	fp := fileLE.Text()
	if fp == "" {
		rep.Log("✗ 请先选择文件")
		return
	}

	op := opCB.Text()

	if strings.HasPrefix(op, "切割") {
		size := sizeCB.Text()
		hashAlgo := "md5"
		if strings.HasPrefix(hashCB.Text(), "sha256") {
			hashAlgo = "sha256"
		}
		outDir := outLE.Text()

		rep.Log("开始切割: %s (大小=%s, 算法=%s)", filepath.Base(fp), size, hashAlgo)
		err := engine.Split(engine.SplitOptions{
			SourcePath: fp,
			ChunkSize:  size,
			OutputDir:  outDir,
			HashAlgo:   hashAlgo,
			ToolVer:    version,
		}, rep)
		if err != nil {
			rep.Log("✗ 错误: %v", err)
		} else {
			rep.Log("✓ 切割完成!")
		}
	} else if strings.HasPrefix(op, "合并") {
		mode := "quick"
		if strings.HasPrefix(modeCB.Text(), "verify") {
			mode = "verify"
		}
		force := forceCB.Checked()
		outDir := outLE.Text()

		rep.Log("开始合并: %s (模式=%s)", filepath.Base(fp), mode)
		err := engine.Merge(engine.MergeOptions{
			ManifestPath: fp,
			Mode:         mode,
			OutputDir:    outDir,
			Force:        force,
		}, rep)
		if err != nil {
			rep.Log("✗ 错误: %v", err)
		} else {
			rep.Log("✓ 合并完成!")
		}
	} else if strings.HasPrefix(op, "校验") {
		rep.Log("开始校验: %s", filepath.Base(fp))
		err := engine.Verify(engine.VerifyOptions{
			ManifestPath: fp,
		}, rep)
		if err != nil {
			rep.Log("✗ 错误: %v", err)
		} else {
			rep.Log("✓ 校验通过!")
		}
	}
}
