@echo off
rem Installs the latest ndo release for Windows, for cmd.exe users who
rem don't want to type a PowerShell one-liner. PowerShell users can just
rem run install.ps1 directly (or the one-liner documented in the README);
rem this is a thin wrapper around the exact same script.
rem
rem   Download and double-click, or from cmd.exe:  install.cmd
rem
rem Override the install directory with NDO_INSTALL_DIR before running.

powershell -NoProfile -ExecutionPolicy Bypass -Command "irm https://raw.githubusercontent.com/green-threads/ndo/main/install/install.ps1 | iex"
