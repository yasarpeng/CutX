using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Input;
using Microsoft.UI.Xaml.Controls;
using Windows.Storage.Pickers;
using WinRT.Interop;
using CutX.GUI;

namespace CutX.GUI.Views;

public sealed partial class MergePage : Page
{
    private string? _manifestFile;

    public MergePage() => InitializeComponent();

    private void ManifestDrop_DragOver(object sender, DragEventArgs e)
    {
        e.AcceptedOperation = Windows.ApplicationModel.DataTransfer.DataPackageOperation.Copy;
    }

    private async void ManifestDrop_Drop(object sender, DragEventArgs e)
    {
        var def = e.DataView.GetDeferral();
        var files = await e.DataView.GetStorageItemsAsync();
        def.Complete();
        if (files.Count > 0)
            SetManifestFile(files[0].Path);
    }

    private async void ManifestDrop_Tapped(object sender, TappedRoutedEventArgs e)
    {
        var picker = new FileOpenPicker();
        InitializeWithWindow.Initialize(picker,
            Win32Interop.GetWindowFromWindowId(App.MainWindow!.AppWindow.Id));
        picker.FileTypeFilter.Add(".json");
        var file = await picker.PickSingleFileAsync();
        if (file != null)
            SetManifestFile(file.Path);
    }

    private void SetManifestFile(string path)
    {
        _manifestFile = path;
        ManifestFileText.Text = path;
    }

    private async void Merge_Click(object sender, RoutedEventArgs e)
    {
        if (_manifestFile == null)
        {
            await ShowDialog("提示", "请先选择清单文件");
            return;
        }

        MergeButton.IsEnabled = false;

        var mode = MergeModeCombo.SelectedItem as string == "verify (边写边校验)" ? "verify" : "quick";
        var outDir = MergeOutputText.Text;
        var force = ForceOverwriteCheck.IsChecked == true ? "--force" : "";

        var args = $"merge \"{_manifestFile}\" --mode {mode} {force}";
        if (!string.IsNullOrWhiteSpace(outDir) && !outDir.Contains("清单"))
            args += $" -o \"{outDir}\"";

        var svc = (App.MainWindow as MainWindow)!.ProcessService;
        await svc.RunAsync(args);

        MergeButton.IsEnabled = true;
    }

    private async Task ShowDialog(string title, string content)
    {
        await new ContentDialog
        {
            Title = title, Content = content, CloseButtonText = "确定", XamlRoot = XamlRoot
        }.ShowAsync();
    }
}
