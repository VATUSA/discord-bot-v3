package config

import "os"
import _ "github.com/joho/godotenv/autoload"

var VATUSA_API2_URL = os.Getenv("VATUSA_API2_URL")
var CONFIG_PATH = os.Getenv("CONFIG_PATH")
