# Installs the latest ndo release for Windows.
#
#   irm https://raw.githubusercontent.com/green-threads/ndo/main/install/install.ps1 | iex
#
# Override the install directory with $env:NDO_INSTALL_DIR (default:
# $env:LOCALAPPDATA\Programs\ndo).

$ErrorActionPreference = "Stop"

$repo = "green-threads/ndo"

$arch = $env:PROCESSOR_ARCHITECTURE
if ($arch -ne "AMD64") {
    Write-Error "ndo: unsupported architecture: $arch (only amd64 releases are published for Windows)"
    exit 1
}
$arch = "amd64"

$version = $env:NDO_VERSION
if (-not $version) {
    $release = Invoke-RestMethod -UseBasicParsing "https://api.github.com/repos/$repo/releases/latest"
    $version = $release.tag_name
    if (-not $version) {
        Write-Error "ndo: could not resolve the latest release tag"
        exit 1
    }
}

$versionNoV = $version.TrimStart("v")
$url = "https://github.com/$repo/releases/download/$version/ndo_${versionNoV}_windows_${arch}.zip"

$installDir = $env:NDO_INSTALL_DIR
if (-not $installDir) {
    $installDir = Join-Path $env:LOCALAPPDATA "Programs\ndo"
}
New-Item -ItemType Directory -Force -Path $installDir | Out-Null

$tmp = Join-Path ([System.IO.Path]::GetTempPath()) ([System.IO.Path]::GetRandomFileName())
New-Item -ItemType Directory -Force -Path $tmp | Out-Null
try {
    $zipPath = Join-Path $tmp "ndo.zip"
    Write-Host "ndo: downloading $url"
    Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $zipPath

    Expand-Archive -Path $zipPath -DestinationPath $tmp -Force
    Copy-Item -Path (Join-Path $tmp "ndo.exe") -Destination (Join-Path $installDir "ndo.exe") -Force
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}

Write-Host "ndo: installed to $installDir\ndo.exe"

$pathDirs = $env:Path -split ";"
if ($pathDirs -notcontains $installDir) {
    Write-Warning "ndo: $installDir is not on your PATH. Add it, e.g.:`n  [Environment]::SetEnvironmentVariable('Path', `"`$env:Path;$installDir`", 'User')"
}

& (Join-Path $installDir "ndo.exe") --version
