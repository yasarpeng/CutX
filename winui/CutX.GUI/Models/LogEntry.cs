using Microsoft.UI.Xaml;
using Microsoft.UI.Xaml.Media;

namespace CutX.GUI.Models;

/// <summary>
/// A single log line displayed in the log panel.
/// Level determines the foreground color.
/// </summary>
public class LogEntry
{
    public string Timestamp { get; set; } = "";
    public string Message { get; set; } = "";
    public LogLevel Level { get; set; } = LogLevel.Info;

    public string DisplayText => $"[{Timestamp}] {Message}";

    public Brush LevelBrush
    {
        get
        {
            var res = Application.Current.Resources;
            return Level switch
            {
                LogLevel.Success => (Brush)res["SuccessBrush"],
                LogLevel.Warning => (Brush)res["WarningBrush"],
                LogLevel.Error   => (Brush)res["ErrorBrush"],
                _                => (Brush)res["TextSecondaryBrush"],
            };
        }
    }
}

public enum LogLevel
{
    Info,
    Success,
    Warning,
    Error,
}
