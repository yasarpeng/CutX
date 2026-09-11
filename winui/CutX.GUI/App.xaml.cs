using Microsoft.UI.Xaml;

namespace CutX.GUI;

public partial class App : Application
{
    public static Window? MainWindow { get; private set; }

    public App()
    {
        InitializeComponent();
    }

    protected override void OnLaunched(LaunchActivatedEventArgs args)
    {
        MainWindow = new MainWindow();
        // Force dark theme
        if (MainWindow.Content is FrameworkElement root)
            root.RequestedTheme = ElementTheme.Dark;

        MainWindow.Activate();
    }
}
