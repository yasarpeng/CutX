//go:build windows

package app

import (
	"fmt"
	"path/filepath"
	"sync/atomic"
	"time"

	"github.com/pengyongshi/cutx/internal/chunk"
	"github.com/pengyongshi/cutx/internal/engine"

	"github.com/lxn/walk"
	. "github.com/lxn/walk/declarative"
)

// guiReporter implements engine.Reporter for the Walk GUI.
type guiReporter struct {
	syncFunc    func(f func())
	progressBar *walk.ProgressBar
	logView     *walk.TextEdit
	label       *walk.Label
	total       int64
	done        int64
	startTime   time.Time
}

func (r *guiReporter) Log(format string, args ...interface{}) {
	msg := fmt.Sprintf("[cutx] "+format, args...)
	r.syncFunc(func() {
		if r.logView != nil {
			r.logView.AppendText(msg + "\r\n")
		}
	})
}

func (r *guiReporter) ProgressStart(total int64, label string) {
	atomic.StoreInt64(&r.total, total)
	atomic.StoreInt64(&r.done, 0)
	r.startTime = time.Now()
	r.syncFunc(func() {
		if r.progressBar != nil {
			r.progressBar.SetRange(0, int(total))
			r.progressBar.SetValue(0)
		}
		if r.label != nil {
			r.label.SetText(label)
		}
	})
}

func (r *guiReporter) ProgressAdd(n int64) {
	cur := atomic.AddInt64(&r.done, n)
	r.syncFunc(func() {
		if r.progressBar != nil {
			r.progressBar.SetValue(int(cur))
		}
	})
}

func (r *guiReporter) ProgressFinish() {
	r.syncFunc(func() {
		if r.progressBar != nil {
			r.progressBar.SetValue(int(atomic.LoadInt64(&r.total)))
		}
	})
}

func (r *guiReporter) ProgressStop() {}

var _ engine.Reporter = (*guiReporter)(nil)

var running int32 // atomic flag: 0=idle, 1=running

// Run starts the GUI application.
func Run() {
	var mw *walk.MainWindow
	var pb *walk.ProgressBar
	var logTE *walk.TextEdit
	var statusLbl *walk.Label

	// Split tab controls
	var splitFileEdit *walk.LineEdit
	var splitSizeEdit *walk.LineEdit
	var splitHashCB *walk.ComboBox
	var splitOutputEdit *walk.LineEdit
	var splitPreviewLbl *walk.Label

	// Merge tab controls
	var mergeFileEdit *walk.LineEdit
	var mergeModeCB *walk.ComboBox
	var mergeOutputEdit *walk.LineEdit
	var mergeForceCB *walk.CheckBox

	// Verify tab controls
	var verifyFileEdit *walk.LineEdit

	newReporter := func() *guiReporter {
		return &guiReporter{
			syncFunc:    mw.Synchronize,
			progressBar: pb,
			logView:     logTE,
			label:       statusLbl,
		}
	}

	runAsync := func(fn func(rep *guiReporter) error) {
		if !atomic.CompareAndSwapInt32(&running, 0, 1) {
			walk.MsgBox(mw, "提示", "有操作正在进行中...", walk.MsgBoxIconWarning)
			return
		}
		rep := newReporter()
		go func() {
			err := fn(rep)
			atomic.StoreInt32(&running, 0)
			mw.Synchronize(func() {
				if err != nil {
					rep.Log("✗ Error: %v", err)
					walk.MsgBox(mw, "错误", fmt.Sprintf("%v", err), walk.MsgBoxIconError)
				} else {
					walk.MsgBox(mw, "完成", "操作成功完成!", walk.MsgBoxIconInformation)
				}
			})
		}()
	}

	// File picker helpers
	pickFile := func(filter string) string {
		dlg := &walk.FileDialog{Filter: filter, Title: "选择文件"}
		if ok, _ := dlg.ShowOpen(mw); ok {
			return dlg.FilePath
		}
		return ""
	}
	pickFolder := func() string {
		dlg := &walk.FileDialog{Title: "选择目录"}
		if ok, _ := dlg.ShowBrowseFolder(mw); ok {
			return dlg.FilePath
		}
		return ""
	}

	// --- Split tab ---
	splitTab := TabPage{
		Title:  "切割 Split",
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "源文件:"},
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 0},
						Children: []Widget{
							LineEdit{AssignTo: &splitFileEdit},
							PushButton{
								Text: "浏览...",
								OnClicked: func() {
									if p := pickFile("All files (*.*)|*.*"); p != "" {
										splitFileEdit.SetText(p)
										updateSplitPreview(splitFileEdit, splitOutputEdit, splitPreviewLbl)
									}
								},
							},
						},
					},
					Label{Text: "切割大小:"},
					LineEdit{AssignTo: &splitSizeEdit, Text: "2G"},
					Label{Text: "校验算法:"},
					ComboBox{AssignTo: &splitHashCB, Value: "md5", Model: []string{"md5", "sha256"}},
					Label{Text: "输出目录:"},
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 0},
						Children: []Widget{
							LineEdit{AssignTo: &splitOutputEdit, CueBanner: "留空=自动创建子目录"},
							PushButton{
								Text: "浏览...",
								OnClicked: func() {
									if p := pickFolder(); p != "" {
										splitOutputEdit.SetText(p)
										updateSplitPreview(splitFileEdit, splitOutputEdit, splitPreviewLbl)
									}
								},
							},
						},
					},
				},
			},
			Label{AssignTo: &splitPreviewLbl, Text: "将创建 cutx-<文件名>/ 子目录"},
			PushButton{
				Text: "▶ 开始切割",
				OnClicked: func() {
					src := splitFileEdit.Text()
					if src == "" {
						walk.MsgBox(mw, "提示", "请选择源文件", walk.MsgBoxIconWarning)
						return
					}
					runAsync(func(rep *guiReporter) error {
						return engine.Split(engine.SplitOptions{
							SourcePath: src,
							ChunkSize:  splitSizeEdit.Text(),
							OutputDir:  splitOutputEdit.Text(),
							HashAlgo:   splitHashCB.Text(),
							ToolVer:     "cutx-gui v1.0.0",
						}, rep)
					})
				},
			},
		},
	}

	// --- Merge tab ---
	mergeTab := TabPage{
		Title:  "合并 Merge",
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "清单文件:"},
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 0},
						Children: []Widget{
							LineEdit{AssignTo: &mergeFileEdit},
							PushButton{
								Text: "浏览...",
								OnClicked: func() {
									if p := pickFile("Manifest files (*.manifest.json)|*.manifest.json|All files (*.*)|*.*"); p != "" {
										mergeFileEdit.SetText(p)
									}
								},
							},
						},
					},
					Label{Text: "合并模式:"},
					ComboBox{AssignTo: &mergeModeCB, Value: "quick (快速)", Model: []string{"quick (快速)", "verify (校验)"}},
					Label{Text: "输出目录:"},
					LineEdit{AssignTo: &mergeOutputEdit, CueBanner: "留空=清单所在目录"},
					Label{Text: ""},
					CheckBox{AssignTo: &mergeForceCB, Text: "强制覆盖已存在的文件"},
				},
			},
			PushButton{
				Text: "▶ 开始合并",
				OnClicked: func() {
					mf := mergeFileEdit.Text()
					if mf == "" {
						walk.MsgBox(mw, "提示", "请选择清单文件", walk.MsgBoxIconWarning)
						return
					}
					mode := "quick"
					if mergeModeCB.Text() == "verify (校验)" {
						mode = "verify"
					}
					outDir := mergeOutputEdit.Text()
					runAsync(func(rep *guiReporter) error {
						return engine.Merge(engine.MergeOptions{
							ManifestPath: mf,
							Mode:         mode,
							OutputDir:    outDir,
							Force:        mergeForceCB.Checked(),
						}, rep)
					})
				},
			},
		},
	}

	// --- Verify tab ---
	verifyTab := TabPage{
		Title:  "校验 Verify",
		Layout: VBox{MarginsZero: true},
		Children: []Widget{
			Composite{
				Layout: Grid{Columns: 2},
				Children: []Widget{
					Label{Text: "清单文件:"},
					Composite{
						Layout: HBox{MarginsZero: true, Spacing: 0},
						Children: []Widget{
							LineEdit{AssignTo: &verifyFileEdit},
							PushButton{
								Text: "浏览...",
								OnClicked: func() {
									if p := pickFile("Manifest files (*.manifest.json)|*.manifest.json|All files (*.*)|*.*"); p != "" {
										verifyFileEdit.SetText(p)
									}
								},
							},
						},
					},
				},
			},
			PushButton{
				Text: "▶ 开始校验",
				OnClicked: func() {
					mf := verifyFileEdit.Text()
					if mf == "" {
						walk.MsgBox(mw, "提示", "请选择清单文件", walk.MsgBoxIconWarning)
						return
					}
					runAsync(func(rep *guiReporter) error {
						return engine.Verify(engine.VerifyOptions{
							ManifestPath: mf,
						}, rep)
					})
				},
			},
		},
	}

	// --- Main window ---
	mw_ := MainWindow{
		AssignTo: &mw,
		Title:    "CutX — 跨平台离线文件切割与合并工具",
		MinSize:  Size{Width: 600, Height: 500},
		Size:     Size{Width: 700, Height: 600},
		Layout:   VBox{MarginsZero: true},
		Children: []Widget{
			TabWidget{Pages: []TabPage{splitTab, mergeTab, verifyTab}},
			Composite{
				Layout: VBox{MarginsZero: true},
				Children: []Widget{
					Label{AssignTo: &statusLbl, Text: "就绪"},
					ProgressBar{AssignTo: &pb, MaxValue: 100},
				},
			},
			GroupBox{
				Title:  "日志输出",
				Layout: VBox{MarginsZero: true},
				Children: []Widget{
					TextEdit{AssignTo: &logTE, ReadOnly: true, VScroll: true, MinSize: Size{Height: 120}},
				},
			},
		},
	}

	if _, err := mw_.Run(); err != nil {
		fmt.Println("GUI error:", err)
	}
}

func updateSplitPreview(fileEdit, outputEdit *walk.LineEdit, previewLbl *walk.Label) {
	src := fileEdit.Text()
	if src == "" {
		previewLbl.SetText("将创建 cutx-<文件名>/ 子目录")
		return
	}
	out := outputEdit.Text()
	if out != "" {
		previewLbl.SetText(fmt.Sprintf("输出到: %s", out))
	} else {
		base := chunk.BaseName(src)
		stripped := chunk.StripExtensions(base)
		dir := filepath.Dir(src)
		previewLbl.SetText(fmt.Sprintf("将创建: %s\\cutx-%s\\", dir, stripped))
	}
}
