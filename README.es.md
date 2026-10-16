# safebox

[English](README.md) | [简体中文](README.zh-CN.md) | [日本語](README.ja.md) | [한국어](README.ko.md) | [Español](README.es.md) | [Français](README.fr.md)

Guía de uso de IA: [English](AI_GUIDE.md) | [简体中文](AI_GUIDE.zh-CN.md)

[![CI](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml/badge.svg)](https://github.com/0xtangmengyao/safebox/actions/workflows/ci.yml)
[![License: MIT](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Go Reference](https://img.shields.io/badge/go-1.25%2B-00ADD8.svg)](go.mod)

`safebox` es una herramienta de línea de comandos pequeña, segura y multiplataforma escrita en Go para cifrar archivos y texto locales. También funciona bien para mover texto cifrado entre aplicaciones mediante una tubería.

- Se apoya únicamente en la madura biblioteca estándar [age](https://age-encryption.org) (`filippo.io/age`). **Sin cifrados ni protocolos propios.**
- El texto cifrado se escribe por defecto como **texto armor** de age (`-----BEGIN AGE ENCRYPTED FILE-----`), fácil de copiar, pegar y canalizar.
- Lee y escribe stdin/stdout por defecto y funciona tanto en shells POSIX como en PowerShell. **El texto en claro nunca se acepta como argumento de línea de comandos**, por lo que no puede filtrarse al historial del shell ni a la lista de procesos.
- Una CLI puramente local: **sin telemetría, sin rastreo y sin ningún tipo de acceso a la red en tiempo de ejecución.** Nunca se conecta a Internet por sí sola.

## Características

- Cifrado y descifrado age/X25519 estándar mediante `filippo.io/age`.
- Orientado a flujos, por lo que los archivos grandes no necesitan caber en memoria.
- Se niega a sobrescribir archivos existentes y usa `O_EXCL` para evitar condiciones de carrera.
- Verifica la etiqueta de autenticación completa antes de escribir cualquier texto en claro.
- Ayuda de la CLI localizada en inglés, chino simplificado, japonés, coreano, español y francés.

## Instalación y compilación

Se requiere Go 1.25 o superior (la versión mínima de la dependencia actual `filippo.io/age`).

```sh
git clone https://github.com/0xtangmengyao/safebox.git safebox
cd safebox
go build -o safebox .
```

Compilación cruzada (Go puro, sin CGO):

```sh
# Linux amd64
GOOS=linux   GOARCH=amd64 go build -o safebox-linux .
# macOS (Apple Silicon)
GOOS=darwin  GOARCH=arm64 go build -o safebox-macos .
# Windows
GOOS=windows GOARCH=amd64 go build -o safebox.exe .
```

Instalar en `$GOBIN` / `%GOPATH%\bin`:

```sh
go install .
```

### Binarios precompilados

El directorio `dist/` publica compilaciones para Linux, macOS y Windows en amd64 y arm64. Elija el archivo correspondiente a su sistema operativo y arquitectura de CPU. `dist/SHA256SUMS` permite comprobar que una descarga está intacta, pero una suma de comprobación por sí sola no demuestra quién la publicó. Los binarios aún no están firmados; para usos muy sensibles, compile desde el código fuente y verifique el origen usted mismo.

Mostrar la versión y la ayuda:

```sh
safebox version
safebox help
```

## Comandos

### Generar un par de claves

```sh
safebox keygen --identity ./keys/alice.txt
```

- La clave privada se escribe en `./keys/alice.txt` solicitando el modo `0600` en sistemas POSIX; la clave pública se escribe en `./keys/alice.txt.pub` solicitando el modo `0644`.
- En sistemas POSIX los **permisos efectivos también están limitados por el umask del proceso**. El umask solo elimina bits de permiso, por lo que el resultado nunca es más amplio de lo solicitado. Por ejemplo, con umask `022` la clave pública es `0644`, y con umask `077` es `0600`. La clave privada solicita `0600`, que no tiene bits de grupo ni otros, por lo que normalmente no se ve afectada por el umask.
- En Windows, el acceso a los archivos lo controla la ACL heredada del directorio padre; asegúrese de que ese directorio solo sea accesible para las cuentas previstas.
- Si cualquiera de los archivos ya existe, `safebox` se niega a continuar. **Nunca sobrescribe.**
- La clave pública (destinatario) también se imprime en stdout para compartirla fácilmente.

### Cifrar

Por defecto, el texto en claro se lee de stdin y el texto cifrado armor se escribe en stdout:

```sh
safebox encrypt --recipient age1... < plain.txt
```

Puede indicar varios destinatarios; cualquier clave privada coincidente puede descifrar:

```sh
safebox encrypt --recipient age1aaa --recipient age1bbb < plain.txt
```

También puede usar rutas de archivo. Los archivos de salida se crean con modo `0600` y **nunca se sobrescriben**:

```sh
safebox encrypt --recipient age1... --in plain.txt --out secret.age
```

En sistemas POSIX el modo de salida solicitado es `0600` (de nuevo limitado por el umask, solo más estrecho, nunca más amplio). En Windows se aplica la ACL heredada del directorio padre.

> La redirección del shell como `> archivo` la gestiona el propio shell y puede truncar un archivo existente antes de que se inicie `safebox`. Para guardar la salida, use `--out PATH` y así conservar la protección contra sobrescritura de `safebox`.

### Descifrar

Por defecto, el texto cifrado se lee de stdin y el texto en claro se escribe en stdout:

```sh
safebox decrypt --identity ./keys/alice.txt < secret.age
```

Forma con archivos (modo de salida POSIX `0600`, ACL heredada en Windows, **nunca sobrescribe**):

```sh
safebox decrypt --identity ./keys/alice.txt --in secret.age --out plain.txt
```

Los archivos de identidad se analizan con las reglas estándar de age: se aceptan archivos estilo age-keygen, se ignoran las líneas vacías y las líneas de comentario `#` (como `# created: ...` y `# public key: age1...`), y un mismo archivo puede contener **varias identidades**; cualquiera que coincida puede descifrar.

El descifrado primero vuelca el texto en claro en un archivo temporal con permisos restringidos en el directorio temporal del sistema y solo lo escribe cuando se ha verificado por completo la etiqueta de autenticación. Si la autenticación falla, no se emite ningún texto en claro y el uso de memoria no crece con el tamaño del archivo. El archivo temporal se elimina al salir, pero **eliminar no es un borrado seguro**: el disco que contiene el directorio temporal, la swap o las copias de seguridad aún pueden retener datos. Se aceptan tanto texto armor como entrada age binaria.

Si `--out` apunta a una ruta existente, `safebox` se niega antes de descifrar o crear cualquier texto en claro temporal. Esa comprobación de existencia no sigue enlaces simbólicos, por lo que un enlace roto también cuenta como existente y se rechaza; la escritura final sigue usando `O_EXCL` como protección autoritativa. Un destino `--out` existente nunca se sobrescribe ni se elimina. Si esta ejecución creó un archivo nuevo y luego sufrió un fallo de escritura, se limpia la salida parcial de esta ejecución.

## Ejemplos

### Texto por una tubería (Linux / macOS / Git Bash)

```sh
printf 'hello safebox\n' | safebox encrypt --recipient AGE1... | safebox decrypt --identity alice.txt
```

### Archivos

```sh
safebox keygen --identity alice.txt
safebox encrypt --recipient "$(cat alice.txt.pub)" --in report.pdf --out report.pdf.age
safebox decrypt --identity alice.txt --in report.pdf.age --out report.pdf
```

### Transferir texto cifrado entre aplicaciones

```sh
# La aplicación A produce texto cifrado en la tubería
producer-app | safebox encrypt --recipient age1... | consumer-app
```

### Windows PowerShell

```powershell
# Generar un par de claves
safebox keygen --identity .\keys\alice.txt

# El modo archivo evita que PowerShell reescriba codificación o saltos de línea
safebox encrypt --recipient (Get-Content -Raw .\keys\alice.txt.pub).Trim() `
    --in .\plain.txt --out .\secret.age

# Descifrar
safebox decrypt --identity .\keys\alice.txt `
    --in .\secret.age --out .\plain.out.txt
```

> Nota: el modo archivo no pasa por la tubería de objetos de texto de PowerShell, por lo que sirve para archivos arbitrarios. Si escribe texto sensible directamente en un comando `echo`/`printf`, el historial del shell podría registrarlo. Introduzca el texto sensible desde un archivo, una herramienta de portapapeles o una aplicación anterior mediante stdin. Nunca ponga texto en claro ni una clave privada en la línea de comandos de `safebox`.

## Idioma y localización

La ayuda y los diagnósticos de la CLI están disponibles en inglés (`en`), chino simplificado (`zh-CN`), japonés (`ja`), coreano (`ko`), español (`es`) y francés (`fr`). El valor predeterminado es el inglés.

Cambie de idioma explícitamente con la opción global `--lang` o la variable de entorno `SAFEBOX_LANG`:

```sh
safebox --lang ja help      # ayuda en japonés
SAFEBOX_LANG=fr safebox decrypt   # diagnósticos en francés
```

`--lang` tiene prioridad sobre `SAFEBOX_LANG`, y los valores desconocidos recurren al inglés. Para obtener ayuda sobre cómo usar tu propio asistente de IA con `safebox`, consulta la [guía de uso de IA](AI_GUIDE.md) ([简体中文](AI_GUIDE.zh-CN.md)).

## Límites de seguridad

- **Criptografía**: age / X25519 estándar mediante `filippo.io/age`. `safebox` no inventa criptografía propia.
- **Claves privadas**: en sistemas POSIX el archivo de clave privada se crea con `0600`; en Windows lo rige la ACL del directorio padre. `safebox` nunca imprime la clave privada. Proteja usted mismo el archivo de clave y no lo guarde en discos compartidos o sin cifrar.
- **Texto en claro temporal**: el descifrado usa un archivo temporal con permisos restringidos para verificar todo el flujo. Se elimina después, pero no se garantiza un borrado seguro en SSD, swap, copias de seguridad o instantáneas. Para alta confidencialidad, mantenga el directorio temporal y la swap en almacenamiento cifrado de confianza.
- **Sin recuperación**: si pierde una clave privada, el texto cifrado **no se puede recuperar**. No hay puerta trasera ni mecanismo de recuperación.
- **Metadatos**: el cifrado protege solo el contenido de los archivos. age no oculta el tamaño, las marcas de tiempo ni los nombres de archivo, y no oculta los patrones de tráfico.
- **Terminales y portapapeles**: copiar claves o texto en claro a través de una terminal puede dejarlos en el historial, el búfer de desplazamiento o el portapapeles.
- **Entropía de frases de contraseña**: esta herramienta usa cifrado de clave pública y no depende de frases de contraseña; no use frases de baja entropía para el cifrado con frase de contraseña.
- **Herramienta local**: `safebox` es una CLI local, sin ninguna función de red ni telemetría en tiempo de ejecución.

## Desarrollo y pruebas

```sh
gofmt -l .
go vet ./...
go test ./...
```

La suite de pruebas cubre la generación de claves y los permisos de archivo (adaptados por plataforma), los idas y vueltas por stdin/stdout y por archivo, múltiples destinatarios, el fallo sin filtración de texto en claro cuando se manipula el texto cifrado o se usa la clave equivocada, y la negativa a sobrescribir archivos existentes. También comprueba la completitud de las traducciones en todos los idiomas publicados.

La integración continua compila y prueba cada push y pull request, ejecuta `gofmt`, `go vet` y `go test`, y compila de forma cruzada para Linux, macOS y Windows sin dejar artefactos en el repositorio.

## Metadatos del proyecto (sugeridos)

Estos topics de GitHub son sugerencias que puede configurar en los ajustes del repositorio; este repositorio no modifica automáticamente los metadatos en línea:

`age`, `encryption`, `cryptography`, `x25519`, `cli`, `golang`, `privacy`, `command-line-tool`, `file-encryption`, `cross-platform`

## Contribuir

Consulte [CONTRIBUTING.md](CONTRIBUTING.md) para saber cómo compilar, probar y enviar cambios. Informe de las vulnerabilidades en privado como se describe en [SECURITY.md](SECURITY.md). Este proyecto sigue el [Contributor Covenant](CODE_OF_CONDUCT.md).

## Licencia

MIT. Consulte [LICENSE](LICENSE).
