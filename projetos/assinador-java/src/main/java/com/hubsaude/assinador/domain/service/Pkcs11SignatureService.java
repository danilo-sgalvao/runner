package com.hubsaude.assinador.domain.service;

import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.domain.model.ValidateRequest;

import java.nio.charset.StandardCharsets;
import java.security.*;
import java.security.cert.Certificate;
import java.util.Base64;
import java.util.Enumeration;

/**
 * Implementação de {@link SignatureService} que delega ao dispositivo criptográfico
 * via PKCS#11 (SunPKCS11 provider). O KeyStore deve ser carregado pelo chamador
 * com {@link Pkcs11ServiceFactory}.
 *
 * <p>Para assinar, o campo {@code token} de {@link SignRequest} identifica o alias
 * da chave privada no KeyStore PKCS#11. Para validar, todos os certificados do
 * dispositivo são tentados até que um confirme a assinatura.
 *
 * <p>Quando o dispositivo não está disponível (biblioteca ausente, PIN errado etc.),
 * a criação do KeyStore falha em {@link Pkcs11ServiceFactory} e o sistema cai de
 * volta para {@link FakeSignatureService}.
 */
public class Pkcs11SignatureService implements SignatureService {

    private final KeyStore keyStore;

    public Pkcs11SignatureService(KeyStore keyStore) {
        this.keyStore = keyStore;
    }

    @Override
    public SignatureResult sign(SignRequest request) {
        try {
            PrivateKey privateKey = (PrivateKey) keyStore.getKey(request.getToken(), null);
            if (privateKey == null) {
                return new SignatureResult(null, false,
                        "Chave privada não encontrada no dispositivo com alias: " + request.getToken());
            }

            Signature sig = Signature.getInstance("SHA256withRSA");
            sig.initSign(privateKey);
            sig.update(request.getContent().getBytes(StandardCharsets.UTF_8));

            return new SignatureResult(
                    Base64.getEncoder().encodeToString(sig.sign()),
                    true,
                    "Assinatura criada com sucesso via PKCS#11"
            );
        } catch (Exception e) {
            return new SignatureResult(null, false, "Erro ao assinar via PKCS#11: " + e.getMessage());
        }
    }

    @Override
    public SignatureResult validate(ValidateRequest request) {
        try {
            byte[] sigBytes     = Base64.getDecoder().decode(request.getSignature());
            byte[] contentBytes = request.getContent().getBytes(StandardCharsets.UTF_8);

            Enumeration<String> aliases = keyStore.aliases();
            while (aliases.hasMoreElements()) {
                Certificate cert = keyStore.getCertificate(aliases.nextElement());
                if (cert == null) continue;
                try {
                    Signature sig = Signature.getInstance("SHA256withRSA");
                    sig.initVerify(cert.getPublicKey());
                    sig.update(contentBytes);
                    if (sig.verify(sigBytes)) {
                        return new SignatureResult(request.getSignature(), true, "Assinatura é válida");
                    }
                } catch (Exception ignored) {
                    // tenta próximo certificado
                }
            }
            return new SignatureResult(request.getSignature(), false, "Assinatura é inválida");

        } catch (IllegalArgumentException e) {
            return new SignatureResult(request.getSignature(), false,
                    "Assinatura com encoding inválido: " + e.getMessage());
        } catch (Exception e) {
            return new SignatureResult(request.getSignature(), false,
                    "Erro ao validar via PKCS#11: " + e.getMessage());
        }
    }
}
