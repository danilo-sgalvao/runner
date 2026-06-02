package com.hubsaude.assinador.infrastructure.pkcs11;

/**
 * Configuração do dispositivo PKCS#11, lida de variáveis de ambiente:
 *   HUBSAUDE_PKCS11_LIBRARY — caminho para a biblioteca nativa (.so/.dll)
 *   HUBSAUDE_PKCS11_NAME    — nome lógico do provider (padrão: HubSaudePKCS11)
 *   HUBSAUDE_PKCS11_PIN     — PIN/senha de acesso ao keystore
 *
 * Retorna {@code null} de {@link #fromEnvironment()} quando a variável
 * HUBSAUDE_PKCS11_LIBRARY não está configurada, indicando que o modo
 * simulado (fake) deve ser usado.
 */
public record Pkcs11Config(String libraryPath, String name, char[] pin) {

    public static Pkcs11Config fromEnvironment() {
        String lib = System.getenv("HUBSAUDE_PKCS11_LIBRARY");
        if (lib == null || lib.isBlank()) {
            return null;
        }
        String name = System.getenv("HUBSAUDE_PKCS11_NAME");
        if (name == null || name.isBlank()) {
            name = "HubSaudePKCS11";
        }
        String pin = System.getenv("HUBSAUDE_PKCS11_PIN");
        return new Pkcs11Config(lib, name, pin != null ? pin.toCharArray() : new char[0]);
    }
}
