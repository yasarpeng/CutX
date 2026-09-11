using System.Runtime.InteropServices;
using Microsoft.UI;
using Microsoft.UI.Windowing;
using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Windows.Graphics;
using CutX.GUI.Services;
using CutX.GUI.Models;

namespace CutX.GUI;

public sealed partial class MainWindow : Window
{
    [DllImport("user32.dll")]
    private static extern uint GetDpiForWindow(IntPtr hWnd);

    private readonly CutxProcessService _processService;

    public MainWindow()
    {
        InitializeComponent();

        // Size the window: 720×640 for a form-based tool
        var hwnd = Win32Interop.GetWindowFromWindowId(AppWindow.Id);
        var scale = GetDpiForWindow(hwnd) / 96.0;
        AppWindow.Resize(new SizeInt32(
            (int)(720 * scale), (int)(640 * scale)));
        AppWindow.Title = "CutX — 跨平台离线文件切割与合并工具";

        _processService = new CutxProcessService();

        // Wire up log callback
        _processService.LogReceived = (entry) =>
        {
            DispatcherQueue.TryEnqueue(() =>
            {
                LogListView.Items.Add(entry);
                if (LogListView.Items.Count > 0)
                    LogListView.ScrollIntoView(LogListView.Items[LogListView.Items.Count - 1]);
            });
        };

        // Wire up progress callback
        _processService.ProgressUpdated = (percent, status) =>
        {
            DispatcherQueue.TryEnqueue(() =>
            {
                ProgressBar.Value = percent;
                ProgressText.Text = status;
            });
        };

        // Wire up completion callback
        _processService.Completed = (error) =>
        {
            DispatcherQueue.TryEnqueue(() =>
            {
                ProgressBar.Value = 100;
                if (error != null)
                {
                    ProgressText.Text = "操作失败";
                }
                else
                {
                    ProgressText.Text = "操作完成";
                    // Reset progress bar after 3 seconds
                    _ = Task.Delay(3000).ContinueWith(_ =>
                    {
                        DispatcherQueue.TryEnqueue(() =>
                        {
                            if (ProgressText.Text == "操作完成")
                            {
                                ProgressBar.Value = 0;
                                ProgressText.Text = "就绪";
                            }
                        });
                    });
                }
            });
        };

        // Start on the Split page
        NavView.SelectedItem = NavView.MenuItems[0];
        NavigateToPage("Split");
    }

    private void NavView_SelectionChanged(NavigationView sender,
        NavigationViewSelectionChangedEventArgs args)
    {
        if (args.SelectedItemContainer is NavigationViewItem item)
            NavigateToPage(item.Tag?.ToString() ?? "Split");
    }

    private void NavigateToPage(string tag)
    {
        switch (tag)
        {
            case "Merge":
                ContentFrame.Navigate(typeof(Views.MergePage));
                break;
            case "Verify":
                ContentFrame.Navigate(typeof(Views.VerifyPage));
                break;
            case "Version":
                ContentFrame.Navigate(typeof(Views.VersionPage));
                break;
            default:
                ContentFrame.Navigate(typeof(Views.SplitPage));
                break;
        }
    }

    public CutxProcessService ProcessService => _processService;

    // ── Log panel controls ──
    private void ClearLog_Click(object sender, RoutedEventArgs e)
    {
        LogListView.Items.Clear();
    }

    private async void ExportLog_Click(object sender, RoutedEventArgs e)
    {
        var picker = new Windows.Storage.Pickers.FileSavePicker();
        picker.SuggestedStartLocation = Windows.Storage.Pickers.PickerLocationId.DocumentsLibrary;
        picker.FileTypeChoices.Add("Text", new[] { ".txt" });
        picker.SuggestedFileName = "cutx-log";

        var hwnd = Win32Interop.GetWindowFromWindowId(AppWindow.Id);
        WinRT.Interop.InitializeWithWindow.Initialize(picker, hwnd);

        var file = await picker.PickSaveFileAsync();
        if (file != null)
        {
            var lines = LogListView.Items
                .OfType<LogEntry>()
                .Select(e => e.DisplayText);
            await Windows.Storage.FileIO.WriteLinesAsync(file, lines);
        }
    }
}
