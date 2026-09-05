rule Suspicious_PowerShell_Download_Execution
{
  strings:
    $download1 = "Invoke-WebRequest" nocase
    $download2 = "DownloadString" nocase
    $download3 = "System.Net.WebClient" nocase
    $exec1 = "IEX" nocase
    $exec2 = "Invoke-Expression" nocase
  condition:
    any of ($download*) and any of ($exec*)
}

rule Suspicious_Windows_Script_Startup
{
  strings:
    $run_key = "Software\\Microsoft\\Windows\\CurrentVersion\\Run" nocase
    $wscript = "WScript.Shell" nocase
    $regwrite = "RegWrite" nocase
  condition:
    $run_key and ($wscript or $regwrite)
}
