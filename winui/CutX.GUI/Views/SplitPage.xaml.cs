using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Windows.Storage.Pickers;
using WinRT.Interop;
using CutX.GUI;

namespace CutX.GUI.Views;

public sealed partial class SplitPage : Page
{
    private string? _selectedFile;

    public SplitPage()
    {
        InitializeComponent();
        ChunkSizeCombo.SelectionChanged += (s, e) => UpdatePreview();
        OutputDirText.TextChanged += (s, e) => UpdatePreview();
    }

    // ── Drag-and-drop handlers ──
    private void DropZone_DragOver(object sender, DragEventArgs e)
    {
        e.AcceptedOperation = Windows.ApplicationModel.DataTransfer.DataPackageOperation.Copy;
        DropZone.BorderBrush = (Windows.UI.Xaml.Media.Brush)App.Current.Resources["AccentBrush"];
        DropZone.Background = new Windows.UI.Xaml.Media.SolidColorBrush(
            Windows.UI.Color.FromArgb(20, 78, 201, 176));
    }

    private void DropZone_DragLeave(object sender, DragEventArgs e)
    {
        ResetDropZoneVisual();
    }

    private void ResetDropZoneVisual()
    {
        DropZone.BorderBrush = (Windows.UI.Xaml.Media.Brush)App.Current.Resources["StrokeBrush"];
        DropZone.Background = new Windows.UI.Xaml.Media.SolidColorBrush(
            Windows.UI.Color.FromArgb(255, 31, 31, 34));
    }

    private async void DropZone_Drop(object sender, DragEventArgs e)
    {
        ResetDropZoneVisual();
        var def = e.DataView.GetDeferral();
        var files = await e.DataView.GetStorageItemsAsync();
        def.Complete();
        if (files.Count > 0)
            SetSelectedFile(files[0].Path);
    }

    // ── Click to browse ──
    private async void DropZone_Tapped(object sender, TappedRoutedEventArgs e)
    {
        var picker = new FileOpenPicker();
        InitializeWithWindow.Initialize(picker,
            Win32Interop.GetWindowFromWindowId(App.MainWindow!.AppWindow.Id));
        picker.FileTypeFilter.Add("*");
        var file = await picker.PickSingleFileAsync();
        if (file != null)
            SetSelectedFile(file.Path);
    }

    // ── State management ──
    private void SetSelectedFile(string path)
    {
        _selectedFile = path;
        SelectedFileText.Text = path;
        UpdatePreview();
    }

    private void UpdatePreview()
    {
        if (_selectedFile == null) return;
        var outDir = OutputDirText.Text;
        if (!string.IsNullOrWhiteSpace(outDir))
        {
            OutputPreviewText.Text = $"输出到: {outDir}";
        }
        else
        {
            var fileName = System.IO.Path.GetFileName(_selectedFile);
            var stripped = StripExtensions(fileName);
            var dir = System.IO.Path.GetDirectoryName(_selectedFile) ?? ".";
            OutputPreviewText.Text = $"将创建: {dir}\\cutx-{stripped}\\";
        }
    }

    private static string StripExtensions(string fileName)
    {
        var result = fileName;
        while (System.IO.Path.HasExtension(result))
            result = System.IO.Path.GetFileNameWithoutExtension(result);
        return string.IsNullOrEmpty(result) ? fileName : result;
    }

    // ── Browse output directory ──
    private async void BrowseOutput_Click(object sender, RoutedEventArgs e)
    {
        var picker = new FolderPicker();
        InitializeWithWindow.Initialize(picker,
            Win32Interop.GetWindowFromWindowId(App.MainWindow!.AppWindow.Id));
        var folder = await picker.PickSingleFolderAsync();
        if (folder != null)
            OutputDirText.Text = folder.Path;
    }

    // ── Execute split ──
    private async void Split_Click(object sender, RoutedEventArgs e)
    {
        if (_selectedFile == null)
        {
            await ShowDialog("提示", "请先选择源文件");
            return;
        }

        SplitButton.IsEnabled = false;

        var size = ChunkSizeCombo.SelectedItem as string ?? "2G";
        if (size == "自定义...")
            size = "2G";

        var hashAlgo = HashAlgoCombo.SelectedItem as string == "sha256 (安全)" ? "sha256" : "md5";
        var outDir = OutputDirText.Text;
        if (!string.IsNullOrWhiteSpace(outDir) && outDir.Contains("自动"))
            outDir = "";

        var args = $"split \"{_selectedFile}\" -s {size} --hash {hashAlgo}";
        if (!string.IsNullOrWhiteSpace(outDir))
            args += $" -o \"{outDir}\"";

        var svc = (App.MainWindow as MainWindow)!.ProcessService;
        await svc.RunAsync(args);

        SplitButton.IsEnabled = true;
    }

    private async Task ShowDialog(string title, string content)
    {
        var dialog = new ContentDialog
        {
            Title = title,
            Content = content,
            CloseButtonText = "确定",
            XamlRoot = XamlRoot
        };
        await dialog.ShowAsync();
    }
}
