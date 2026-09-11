using System.Text;
using System.Diagnostics;
using System.Text.RegularExpressions;
using CutX.GUI.Models;

namespace CutX.GUI.Services;

/// <summary>
/// Wraps the cutx.exe CLI binary as a subprocess.
/// Parses stderr (where cutx writes [ cutx ] messages and progress) into
/// structured log entries and progress updates.
///
/// Output patterns from cutx stderr:
///   1. Full lines:   "[ cutx ] Cleaning residual chunks...\n"
///   2. Progress:     "\r[ cutx ] Splitting: file ████░░░░ 50.0% | 50GB / 100GB | ETA: 3m05s   "
///
/// Progress lines contain █ (U+2588) and/or ░ (U+2591) and a percentage.
/// We split on these patterns.
/// </summary>
public class CutxProcessService
{
    public Action<LogEntry>? LogReceived { get; set; }
    public Action<double, string>? ProgressUpdated { get; set; }
    public Action<string?>? Completed { get; set; }

    public bool IsRunning => _process != null && !_process.HasExited;

    private Process? _process;

    // Match "XX.X%" anywhere in the string
    private static readonly Regex ProgressPercentRegex =
        new(@"(\d+\.?\d*)\s*%", RegexOptions.Compiled);

    /// <summary>
    /// Launch cutx.exe with the given arguments.
    /// The binary is assumed to be in PATH or in the same directory as the GUI.
    /// </summary>
    public async Task RunAsync(string arguments)
    {
        var psi = new ProcessStartInfo
        {
            FileName = ResolveCutxPath(),
            Arguments = arguments,
            UseShellExecute = false,
            RedirectStandardError = true,
            RedirectStandardOutput = true,
            CreateNoWindow = true,
            StandardErrorEncoding = Encoding.UTF8,
        };

        _process = new Process { StartInfo = psi };
        _process.ErrorDataReceived += OnErrorData;
        _process.Start();
        _process.BeginErrorReadLine();

        await _process.WaitForExitAsync();

        var error = _process.ExitCode != 0
            ? $"cutx exited with code {_process.ExitCode}"
            : null;

        Completed?.Invoke(error);
    }

    /// <summary>Cancel the running process.</summary>
    public void Cancel()
    {
        if (_process != null && !_process.HasExited)
        {
            _process.Kill(entireProcessTree: true);
        }
    }

    // ── Private ──

    private static string ResolveCutxPath()
    {
        // Try same directory as the GUI exe, then fall back to PATH
        var exeDir = AppContext.BaseDirectory;
        var localPath = Path.Combine(exeDir, "cutx.exe");
        if (File.Exists(localPath))
            return localPath;
        return "cutx.exe"; // rely on PATH
    }

    private void OnErrorData(object sender, DataReceivedEventArgs e)
    {
        if (e.Data == null) return;

        var raw = e.Data;
        if (string.IsNullOrWhiteSpace(raw))
            return;

        // Check for progress line (contains █ or ░ characters)
        bool isProgress = raw.Contains('\u2588') || raw.Contains('\u2591');

        if (isProgress)
        {
            var match = ProgressPercentRegex.Match(raw);
            if (match.Success && double.TryParse(match.Groups[1].Value, out var pct))
            {
                // Build a clean status text: strip [ cutx ] prefix, progress bar chars
                var statusText = raw;
                // Remove [ cutx ] prefix
                var cutxIdx = statusText.IndexOf("[ cutx ]");
                if (cutxIdx >= 0)
                    statusText = statusText.Substring(cutxIdx + 8).TrimStart();

                // Remove progress bar characters and clean up
                statusText = Regex.Replace(statusText, @"[█░]+", "");
                statusText = Regex.Replace(statusText, @"\s+", " ").Trim();

                ProgressUpdated?.Invoke(pct, statusText);
                return;
            }
        }

        // Regular log line
        var timestamp = DateTime.Now.ToString("HH:mm:ss");
        var message = raw;

        // Strip [ cutx ] prefix
        if (message.StartsWith("[ cutx ]"))
            message = message.Substring(8).TrimStart();

        // Determine log level from message content
        var level = message.StartsWith("✓") ? LogLevel.Success
                   : message.StartsWith("✗") ? LogLevel.Error
                   : message.StartsWith("⚠") ? LogLevel.Warning
                   : message.StartsWith("Tip:") ? LogLevel.Info
                   : LogLevel.Info;

        LogReceived?.Invoke(new LogEntry
        {
            Timestamp = timestamp,
            Message = message,
            Level = level
        });
    }
}
