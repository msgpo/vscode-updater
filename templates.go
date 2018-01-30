package main

import "html/template"

type pkgbuildData struct {
	Name        string
	Description string
	Version     string
	URL         string
	ArchiveHash string
	DesktopHash string
}

var pkgbuildTmpl = template.Must(template.New("PKGBUILD").Parse(`pkgname={{.Name}}
pkgver={{.Version}}
pkgrel=1
pkgdesc="{{.Description}}"
arch=(x86_64)
url="https://code.visualstudio.com/"
license=('custom: commercial')
depends=(fontconfig libxtst gtk2 python cairo alsa-lib nss gcc-libs gvfs libnotify libxss gconf)
source=({{.URL}}
        ${pkgname}.desktop)
sha256sums=('{{.ArchiveHash}}'
            '{{.DesktopHash}}')

package() {
  install -d "${pkgdir}/usr/share/licenses/${pkgname}"
  install -d "${pkgdir}/opt/${pkgname}"
  install -d "${pkgdir}/usr/bin"
  install -d "${pkgdir}/usr/share/applications"
  install -d "${pkgdir}/usr/share/icons"

  install -m644 "${srcdir}/VSCode-linux-x64/resources/app/LICENSE.txt" "${pkgdir}/usr/share/licenses/${pkgname}/LICENSE"
  install -m644 "${srcdir}/VSCode-linux-x64/resources/app/resources/linux/code.png" "${pkgdir}/usr/share/icons/${pkgname}.png"
  install -m644 "${srcdir}/${pkgname}.desktop" "${pkgdir}/usr/share/applications/${pkgname}.desktop"

  cp -r "${srcdir}/VSCode-linux-x64/"* "${pkgdir}/opt/${pkgname}" -R
  ln -s /opt/${pkgname}/bin/{{.Name}} "${pkgdir}"/usr/bin/{{.Name}}
}
`))

type desktopData struct {
	Name        string
	FullName    string
	Description string
	WMClass     string
}

var desktopTmpl = template.Must(template.New("desktop").Parse(`[Desktop Entry]
Name={{.FullName}}
Comment={{.Description}}
GenericName=Text Editor
Icon={{.Name}}
Exec=/usr/bin/{{.Name}} %f
Type=Application
Terminal=false
StartupNotify=true
StartupWMClass={{.WMClass}}
Categories=Development;WebDevelopment;IDE;Utility;TextEditor;
MimeType=text/plain;inode/directory;
Keywords=vscode;
Actions=new-window;

[Desktop Action new-window]
Name=New Window
Icon={{.Name}}
Exec=/usr/bin/{{.Name}} --new-window %f
`))
