package ota

type Config struct {
    SoftwareDirectory string `yaml:"softwareDirectory"`
    Mqtt struct {
        URL string `yaml:"url"`
        Username string `yaml:"username"`
        Password string `yaml:"password"`
        Topics struct {
            Advertise string `yaml:"advertise"`
            Report string `yaml:"report"`
            Upgrade string `yaml:"upgrade"`
        }
    } `yaml:"mqtt"`
}
