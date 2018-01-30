# vscode-updater

# How to use it

Install the package

```
sudo pacman -Sy go git
git clone https://github.com/zzag/vscode-updater.git
cd vscode-updater/arch
makepkg --clean
sudo pacman -U *.pkg.tar.xz
```

Add the `vscode` repo to the `pacman.conf`

```
[vscode]
SigLevel = Never
Server = file://usr/local/repo/vscode
```

Enable the `vscode-updater` service

```
sudo systemctl enable vscode-updater
```

Also, make sure `NetworkManager-wait-online.service` is enabled

```
sudo systemctl enable NetworkManager-wait-online.service
```

Reboot. Wait until `code` and `code-insiders` are built. You can
check log of the vscode-updater by running

```
systemctl status vscode-updater
```

When it's done, install Visual Studio Code

```
sudo pacman -Syu
sudo pacman -S code # or code-insiders
```
