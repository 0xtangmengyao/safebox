package main

// esMessages is the Spanish catalog.
var esMessages = catalog{
	// Global usage.
	"usage.title":       "safebox %s - cifrado de archivos y texto basado en age",
	"usage.header":      "Uso:",
	"usage.synopsis":    "  safebox [--lang CODE] <comando> [opciones]",
	"usage.commands":    "Comandos:",
	"usage.keygen":      "  keygen   Genera una identidad X25519 y su destinatario público",
	"usage.encrypt":     "  encrypt  Cifra stdin (o --in) como texto armor en stdout (o --out)",
	"usage.decrypt":     "  decrypt  Descifra entrada age armor/binaria a stdout (o --out)",
	"usage.version":     "  version  Muestra la versión",
	"usage.help":        "  help     Muestra esta ayuda",
	"usage.commandHelp": `Ejecute "safebox <comando> -h" para ver las opciones específicas.`,
	"usage.plaintext": "El texto en claro nunca se acepta como argumento de línea de comandos: pásalo\n" +
		"  por stdin para no filtrarlo al historial del shell ni a la lista de procesos.",
	"usage.lang": "Idioma: use --lang en|zh-CN|ja|ko|es|fr o la variable de entorno SAFEBOX_LANG.",

	// Diagnostics.
	"diag.missingCommand":            "falta el comando",
	"diag.unknownCommand":            "comando desconocido %q",
	"diag.unexpectedArgument":        "%s: argumento inesperado %q",
	"diag.keygen.requiresIdentity":   "keygen requiere --identity PATH",
	"diag.encrypt.requiresRecipient": "encrypt requiere al menos un --recipient",
	"diag.decrypt.requiresIdentity":  "decrypt requiere --identity PATH",
	"diag.interrupted":               "interrumpido",

	// Per-command usage.
	"keygen.usageLine": "Uso: safebox keygen --identity PATH",
	"keygen.detail": "Escribe PATH (clave privada, 0600) y PATH.pub (clave pública, 0644).\n" +
		"Ninguno de los dos archivos se sobrescribe nunca.",

	"encrypt.usageLine": "Uso: safebox encrypt --recipient AGE1... [--recipient ...] [--in PATH] [--out PATH]",
	"encrypt.detail": "Lee texto en claro de stdin salvo que se indique --in. Escribe texto armor.\n" +
		"El texto en claro no debe pasarse como argumento.",

	"decrypt.usageLine": "Uso: safebox decrypt --identity PATH [--in PATH] [--out PATH]",
	"decrypt.detail": "Acepta entrada age armor o binaria. El texto en claro se verifica por completo\n" +
		"antes de escribir nada, así que un fallo nunca filtra salida parcial.",

	// Flag descriptions.
	"flag.keygen.identity":   "ruta del archivo de clave privada X25519 a crear (0600)",
	"flag.encrypt.recipient": "clave pública del destinatario (age1...); puede repetirse",
	"flag.encrypt.in":        "archivo de entrada (predeterminado stdin)",
	"flag.encrypt.out":       "archivo de salida (predeterminado stdout); nunca se sobrescribe",
	"flag.decrypt.identity":  "ruta del archivo de clave privada X25519",
	"flag.decrypt.in":        "archivo de entrada (predeterminado stdin)",
	"flag.decrypt.out":       "archivo de salida (predeterminado stdout); nunca se sobrescribe",
}
