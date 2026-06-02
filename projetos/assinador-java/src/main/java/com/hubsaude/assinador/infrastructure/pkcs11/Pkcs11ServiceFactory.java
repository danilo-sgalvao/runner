package com.hubsaude.assinador.infrastructure.pkcs11;

import com.hubsaude.assinador.domain.service.Pkcs11SignatureService;
import com.hubsaude.assinador.domain.service.SignatureService;

import java.security.KeyStore;
import java.security.Provider;
import java.security.Security;

/**
 * Cria um {@link Pkcs11SignatureService} conectado ao dispositivo PKCS#11
 * descrito em {@link Pkcs11Config}.
 *
 * <p>Configura o provider SunPKCS11 com a biblioteca nativa indicada e abre
 * o KeyStore do dispositivo. Lança exceção (que o chamador deve tratar com
 * fallback para {@link com.hubsaude.assinador.domain.service.FakeSignatureService})
 * quando o provider não está disponível ou o dispositivo não responde.
 *
 * <p>Setup para testes com SoftHSM2:
 * <pre>
 *   softhsm2-util --init-token --slot 0 --label "hubsaude" --pin 1234 --so-pin 1234
 *   export HUBSAUDE_PKCS11_LIBRARY=/usr/lib/softhsm/libsofthsm2.so
 *   export HUBSAUDE_PKCS11_PIN=1234
 * </pre>
 */
public class Pkcs11ServiceFactory {

    public static SignatureService create(Pkcs11Config config) throws Exception {
        Provider provider = Security.getProvider("SunPKCS11");
        if (provider == null) {
            throw new IllegalStateException("Provedor SunPKCS11 não disponível nesta plataforma");
        }

        String pkcs11Cfg = String.format("--name=%s%nlibrary=%s%n", config.name(), config.libraryPath());
        provider = provider.configure(pkcs11Cfg);
        Security.addProvider(provider);

        KeyStore ks = KeyStore.getInstance("PKCS11", provider);
        ks.load(null, config.pin());

        return new Pkcs11SignatureService(ks);
    }
}
