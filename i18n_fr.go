package main

// frMessages is the French catalog.
var frMessages = catalog{
	// Global usage.
	"usage.title":       "safebox %s - chiffrement de fichiers et de texte basé sur age",
	"usage.header":      "Utilisation :",
	"usage.synopsis":    "  safebox [--lang CODE] <commande> [options]",
	"usage.commands":    "Commandes :",
	"usage.keygen":      "  keygen   Génère une identité X25519 et son destinataire public",
	"usage.encrypt":     "  encrypt  Chiffre stdin (ou --in) en texte armor vers stdout (ou --out)",
	"usage.decrypt":     "  decrypt  Déchiffre une entrée age armor/binaire vers stdout (ou --out)",
	"usage.version":     "  version  Affiche la version",
	"usage.help":        "  help     Affiche cette aide",
	"usage.commandHelp": `Lancez « safebox <commande> -h » pour les options propres à la commande.`,
	"usage.plaintext": "Le texte en clair n'est jamais accepté comme argument de ligne de commande : " +
		"transmettez-le\n  via stdin pour éviter toute fuite dans l'historique du shell et la liste des processus.",
	"usage.lang": "Langue : utilisez --lang en|zh-CN|ja|ko|es|fr ou la variable d'environnement SAFEBOX_LANG.",

	// Diagnostics.
	"diag.missingCommand":            "commande manquante",
	"diag.unknownCommand":            "commande inconnue %q",
	"diag.unexpectedArgument":        "%s : argument inattendu %q",
	"diag.keygen.requiresIdentity":   "keygen nécessite --identity PATH",
	"diag.encrypt.requiresRecipient": "encrypt nécessite au moins un --recipient",
	"diag.decrypt.requiresIdentity":  "decrypt nécessite --identity PATH",
	"diag.interrupted":               "interrompu",

	// Per-command usage.
	"keygen.usageLine": "Utilisation : safebox keygen --identity PATH",
	"keygen.detail": "Écrit PATH (clé privée, 0600) et PATH.pub (clé publique, 0644).\n" +
		"Aucun des deux fichiers n'est jamais écrasé.",

	"encrypt.usageLine": "Utilisation : safebox encrypt --recipient AGE1... [--recipient ...] [--in PATH] [--out PATH]",
	"encrypt.detail": "Lit le texte en clair depuis stdin sauf si --in est fourni. Écrit du texte armor.\n" +
		"Le texte en clair ne doit pas être passé en argument.",

	"decrypt.usageLine": "Utilisation : safebox decrypt --identity PATH [--in PATH] [--out PATH]",
	"decrypt.detail": "Accepte une entrée age armor ou binaire. Le texte en clair est intégralement vérifié\n" +
		"avant toute écriture : un échec ne divulgue donc jamais de sortie partielle.",

	// Flag descriptions.
	"flag.keygen.identity":   "chemin du fichier de clé privée X25519 à créer (0600)",
	"flag.encrypt.recipient": "clé publique du destinataire (age1...) ; peut être répétée",
	"flag.encrypt.in":        "fichier d'entrée (stdin par défaut)",
	"flag.encrypt.out":       "fichier de sortie (stdout par défaut) ; jamais écrasé",
	"flag.decrypt.identity":  "chemin du fichier de clé privée X25519",
	"flag.decrypt.in":        "fichier d'entrée (stdin par défaut)",
	"flag.decrypt.out":       "fichier de sortie (stdout par défaut) ; jamais écrasé",
}
