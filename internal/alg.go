package mpw

import (
	"crypto"
	"crypto/hmac"
	"encoding/binary"
	"fmt"
	"strings"

	"golang.org/x/crypto/scrypt"
)

func Password(fullName, masterPassword, siteName  string, siteCounter int, resultType ResultType) (string, error) {
	master, err := masterKey(masterPassword, fullName)
	if err != nil {
		return "", err
	}
	site, err := siteKey(siteName, master, uint(siteCounter))
	if err != nil {
		return "", err
	}
	return password(site, resultType)
}

func Identicon(fullName, masterPassword string, useColor bool) string {
	hash := hmac.New(crypto.SHA256.New, []byte(masterPassword))
	hash.Write([]byte(fullName))
	seed := hash.Sum(nil)
	leftArm := leftArms[seed[0] % byte(len(leftArms))]
	body := bodies[seed[1] % byte(len(bodies))]
	rightArm := rightArms[seed[2] % byte(len(rightArms))]
	accessory := accessories[seed[3] % byte(len(accessories))]
	color := colors[seed[4] % byte(len(colors))]
	icon := fmt.Sprintf("%s%s%s%s", leftArm, body, rightArm, accessory)
	if useColor  {
		return fmt.Sprintf("%s%s%s", colorCodes[color], icon, colorCodes["White"])
	}
	return icon
}

// Phase 1: Your identity
//
//  Your identity is defined by your master key.  This key unlocks all of your
//  doors.  Your master key is the cryptographic result of two components:
//
//  1.Your <name> (identification)
//  2.Your <master password> (authentication)
//
//  Your master password is your personal secret and your name scopes that
//  secret to your identity.  Together, they create a cryptographic identifier
//  that is unique to your person.
//
//  ´´´
//  masterKey = SCRYPT( key, seed, N, r, p, dkLen )
//  key = <master password>
//  seed = scope . LEN(<name>) . <name>
//  N = 32768
//  r = 8
//  p = 2
//  dkLen = 64
//  ´´´
//
//  We employ the SCRYPT cryptographic function to derive a 64-byte
//  cryptographic key from the user’s name and master password using a fixed
//  set of parameters.
func masterKey(masterPassword string, name string) ([]byte, error) {
	lengthNameAsBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthNameAsBytes, uint32(len([]byte(name))))
	seed := []byte("com.lyndir.masterpassword")
	seed = append(seed, lengthNameAsBytes...)
	seed = append(seed, []byte(name)...)
	return scrypt.Key([]byte(masterPassword), seed, 32768, 8, 2, 64)	
}

// Phase 2: Your site key "com.lyndir.masterpassword"
// 
// Your site key is a derivative from your master key when it is used to
// unlock the door to a specific site. Your site key is the result of two
// components:
// 
// 1.Your <site name> (identification)
// 2.Your <masterkey> (authentication)
// 3.Your <site counter>
// 
// Your master key ensures only your identity has access to this key and your
// site name scopes the key to your site.  The site counter ensures you can
// easily create new keys for the site should a key become
// compromised. Together, they create a cryptographic identifier that is
// unique to your account at this site.
// 
// siteKey = HMAC-SHA-256( key, seed )
// key = <master key>
// seed = scope . LEN(<site name>) . <site name> . <counter>
// 
// We employ the HMAC-SHA-256 cryptographic function to derive a 64-byte
// cryptographic site key from the from the site name and master key scoped
// to a given counter value.
func siteKey(siteName string, masterKey []byte, counter uint) ([]byte, error) {
	lengthNameAsBytes := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthNameAsBytes, uint32(len([]byte(siteName))))
	counterAsbytes := make([]byte, 4)
	binary.BigEndian.PutUint32(counterAsbytes, uint32(counter))
	seed := []byte("com.lyndir.masterpassword")
	seed = append(seed, lengthNameAsBytes...)
	seed = append(seed, []byte(siteName)...)
	seed = append(seed, counterAsbytes...)

	hash := hmac.New(crypto.SHA256.New, masterKey)
	_ , err := hash.Write(seed)
	if err != nil {
		return nil, err
	}
	return hash.Sum(nil), nil
}

// Output Templates:
// In an effort to enforce increased password entropy, a common consensus has
// developed among account administrators that passwords should adhere to
// certain arbitrary password policies.  These policies enforce certain rules
// which must be honoured for an account password to be deemed acceptable.
//
// As a result of these enforcement practices, Master Password’s site key
// output must necessarily adhere to these types of policies.  Since password
// policies are governed by site administrators and not standardized, Master
// Password defines several password templates to make a best-effort attempt at
// generating site passwords that conform to these policies while also keeping
// its output entropy as high as possible under the constraints.
type ResultType string
type templaceCharacter rune
type template []templaceCharacter
var TemplateDictionary map[ResultType][]template = map[ResultType][]template{
	"Maximum": {
		{'a','n','o','x','x','x','x','x','x','x','x','x','x','x','x','x','x','x','x','x'},
		{'a','x','x','x','x','x','x','x','x','x','x','x','x','x','x','x','x','x','n','o'},
	},
	"Long": {
		{'C','v','c','v','n','o','C','v','c','v','C','v','c','v'},
		{'C','v','c','v','C','v','c','v','n','o','C','v','c','v'},
		{'C','v','c','v','C','v','c','v','C','v','c','v','n','o'},
		{'C','v','c','c','n','o','C','v','c','v','C','v','c','v'},
		{'C','v','c','c','C','v','c','v','n','o','C','v','c','v'},
		{'C','v','c','c','C','v','c','v','C','v','c','v','n','o'},
		{'C','v','c','v','n','o','C','v','c','c','C','v','c','v'},
		{'C','v','c','v','C','v','c','c','n','o','C','v','c','v'},
		{'C','v','c','v','C','v','c','c','C','v','c','v','n','o'},
		{'C','v','c','v','n','o','C','v','c','v','C','v','c','c'},
		{'C','v','c','v','C','v','c','v','n','o','C','v','c','c'},
		{'C','v','c','v','C','v','c','v','C','v','c','c','n','o'},
		{'C','v','c','c','n','o','C','v','c','c','C','v','c','v'},
		{'C','v','c','c','C','v','c','c','n','o','C','v','c','v'},
		{'C','v','c','c','C','v','c','c','C','v','c','v','n','o'},
		{'C','v','c','v','n','o','C','v','c','c','C','v','c','c'},
		{'C','v','c','v','C','v','c','c','n','o','C','v','c','c'},
		{'C','v','c','v','C','v','c','c','C','v','c','c','n','o'},
		{'C','v','c','c','n','o','C','v','c','v','C','v','c','c'},
		{'C','v','c','c','C','v','c','v','n','o','C','v','c','c'},
		{'C','v','c','c','C','v','c','v','C','v','c','c','n','o'},
	},
	"Medium": {
		{'C','v','c','n','o','C','v','c'},
		{'C','v','c','C','v','c','n','o'},
	},
	"Short": {
		{'C','v','c','n'},
	},
	"Basic": {
		{'a','a','a','n','a','a','a','n'},
		{'a','a','n','n','a','a','a','n'},
		{'a','a','a','n','n','a','a','a'},
	},
	"PIN": {
		{'n','n','n','n'},
	},
	"Name": {{'c','v','c','c','v','c','v','c','v'}},
	"Phrase": {
		{'c','v','c','c',' ','c','v','c',' ','c','v','c','c','v','c','v',' ','c','v','c'},
		{'c','v','c',' ','c','v','c','c','v','c','v','c','v',' ','c','v','c','v'},
		{'c','v',' ','c','v','c','c','v',' ','c','v','c',' ','c','v','c','v','c','c','v'},
	},
}
var templateCharsDictionary map[templaceCharacter][]string = map[templaceCharacter][]string{
	'V': {"A","E","I","O","U"},
	'C': {"B","C","D","F","G","H","J","K","L","M","N","P","Q","R","S","T","V","W","X","Y","Z"},
	'v': {"a","e","i","o","u"},
	'c': {"b","c","d","f","g","h","j","k","l","m","n","p","q","r","s","t","v","w","x","y","z"},
	'A': {"A","E","I","O","U","B","C","D","F","G","H","J","K","L","M","N","P","Q","R","S","T","V","W","X","Y","Z"},
	'a': {"A","E","I","O","U","a","e","i","o","u","B","C","D","F","G","H","J","K","L","M","N","P","Q","R","S","T","V","W","X","Y","Z","b","c","d","f","g","h","j","k","l","m","n","p","q","r","s","t","v","w","x","y","z"},
	'n': {"0","1","2","3","4","5","6","7","8","9"},
	'o': {"@","&","%","?",",","=","[","]","_",":","-","+","*","$","#","!","'","^","~",";","(",")","/","."},
	'x': {"A","E","I","O","U","a","e","i","o","u","B","C","D","F","G","H","J","K","L","M","N","P","Q","R","S","T","V","W","X","Y","Z","b","c","d","f","g","h","j","k","l","m","n","p","q","r","s","t","v","w","x","y","z","0","1","2","3","4","5","6","7","8","9","!","@","#","$","%","^","&","*","(",")"},
	' ': {" "},
}

// Phase 3: Your site password
// Your site password is an identifier derived from your site key in
// compoliance with the site’s password policy.
// 
// The purpose of this step is to render the site’s cryptographic key into a
// format that the site’s password input will accept.
// 
// Master Password declares several site password formats and uses these
// pre-defined password “templates” to render the site key legible.
// 
// ´´´
// template = templates[ <site key>[0] % LEN( templates ) ]
// 
// for i in 0..LEN( template )
//   passChars = templateChars[ template[i] ]2
//   passWord[i] = passChars[ <site key>[i+1] % LEN( passChars ) ]
// 
// We resolve a template to use for the password from the site key’s first
// byte.  As we iterate the template, we use it to translate site key bytes
// into password characters.  The result is a site password in the form
// defined by the site template scoped to our site key.
// 
// This password is then used to authenticate the user for his account at
// this site.
func password(siteKey []byte, class ResultType) (string, error) {
	templates, found := TemplateDictionary[class]
	if !found {
		return "", fmt.Errorf("class %s not found", class)
	}
	if len(templates) > 255 {
		return "", fmt.Errorf("template class %s to large, len %d but max 255", class, len(templates))
	}
	template := templates[uint8(siteKey[0]) % uint8(len(templates))]
	if len(template) > len(siteKey) {
		return "", fmt.Errorf("template %s to large, len %d but max 255", class, len(template))
	}
	password := make([]string, 0)
	for i, tc := range template {
		passChars := templateCharsDictionary[tc]
		password = append(password,
			passChars[uint8(siteKey[i+1]) % uint8(len(passChars))],
		)
	}
	return strings.Join(password, ""), nil
}

var leftArms  = []string{"╔", "╚", "╰", "═"}
var rightArms = []string{"╗", "╝", "╯", "═"}
var bodies = []string{"█", "░", "▒", "▓", "☺", "☻"}
var accessories = []string{
    "◈", "◎", "◐", "◑", "◒", "◓", "☀", "☁", "☂", "☃", "", "★",
    "☆", "☎", "☏", "⎈", "⌂", "☘", "☢", "☣", "☕", "⌚", "⌛", "⏰",
    "⚡", "⛄", "⛅", "☔", "♔", "♕", "♖", "♗", "♘", "♙", "♚", "♛",
    "♜", "♝", "♞", "♟", "♨", "♩", "♪", "♫", "⚐", "⚑", "⚔", "⚖",
	"⚙", "⚠", "⌘", "⏎", "✄", "✆", "✈", "✉", "✌"}
type color string
type colorCode string
var colorCodes = map[color]colorCode {
	"Red": "\033[31m",
		"Green": "\033[32m",
		"Yellow":"\033[33m",
		"Blue": "\033[34m",
		"Magenta": "\033[35m",
		"Cyan": "\033[36m",
		"White": "\033[37m",
	}
var colors = []color{"Red", "Green", "Yellow", "Blue", "Magenta", "Cyan", "White"}

