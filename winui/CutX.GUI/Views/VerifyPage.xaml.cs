using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Controls;
using Windows.Storage.Pickers;
using WinRT.Interop;
using CutX.GUI;

namespace CutX.GUI.Views;

public sealed partial class VerifyPage : Page
{
    private string? _manifestFile;

    public VerifyPage() => InitializeComponent();

    private void VerifyDrop_DragOver(object sender, DragEventArgs e)
    {
        e.AcceptedOperation = Windows.ApplicationModel.DataTransfer.DataPackageOperation.Copy;
    }

    private async void VerifyDrop_Drop(object sender, DragEventArgs e)
    {
        var def = e.DataView.GetDeferral();
        var files = await e.DataView.GetStorageItemsAsync();
        def.Complete();
        if (files.Count > 0)
            SetManifestFile(files[0].Path);
    }

    private async void VerifyDrop_Tapped(object sender, TappedRoutedEventArgs e)
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
        VerifyFileText.Text = path;
    }

    private async void Verify_Click(object sender, RoutedEventArgs e)
    {
        if (_manifestFile == null)
        {
            await new ContentDialog
            {
                Title = "提示", Content = "请先选择清单文件",
                CloseButtonText = "确定", XamlRoot = XamlRoot
            }.ShowAsync();
            return;
        }

        VerifyButton.IsEnabled = false;

        var args = $"verify \"{_manifestFile}\"";
        var svc = (App.MainWindow as MainWindow)!.ProcessService;
        await svc.RunAsync(args);

        VerifyButton.IsEnabled = true;
    }
}
