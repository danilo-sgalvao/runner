package com.hubsaude.assinador.domain.service;

import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.domain.model.ValidateRequest;
import org.junit.jupiter.api.Test;

import java.security.KeyStore;
import java.util.Base64;
import java.util.Collections;

import static org.junit.jupiter.api.Assertions.*;
import static org.mockito.Mockito.*;

/**
 * Testa o comportamento de {@link Pkcs11SignatureService} sem dispositivo real.
 * Usa um KeyStore mockado para cobrir caminhos de erro e comportamentos limítrofes.
 * O fluxo completo sign→validate com chave real requer SoftHSM2 instalado
 * e é coberto por testes de integração externos.
 */
class Pkcs11SignatureServiceTest {

    @Test
    void sign_chaveNaoEncontrada_retornaErroComAlias() throws Exception {
        KeyStore ks = mock(KeyStore.class);
        when(ks.getKey("alias-inexistente", null)).thenReturn(null);

        Pkcs11SignatureService service = new Pkcs11SignatureService(ks);
        SignRequest req = new SignRequest();
        req.setContent("documento");
        req.setToken("alias-inexistente");

        SignatureResult result = service.sign(req);

        assertFalse(result.isValid());
        assertNull(result.getSignature());
        assertTrue(result.getMessage().contains("alias-inexistente"));
    }

    @Test
    void sign_excecaoNoKeyStore_retornaErro() throws Exception {
        KeyStore ks = mock(KeyStore.class);
        when(ks.getKey(any(), any())).thenThrow(new RuntimeException("HSM offline"));

        Pkcs11SignatureService service = new Pkcs11SignatureService(ks);
        SignRequest req = new SignRequest();
        req.setContent("documento");
        req.setToken("minha-chave");

        SignatureResult result = service.sign(req);

        assertFalse(result.isValid());
        assertTrue(result.getMessage().contains("PKCS#11"));
    }

    @Test
    void validate_keyStoreVazio_retornaInvalida() throws Exception {
        KeyStore ks = mock(KeyStore.class);
        when(ks.aliases()).thenReturn(Collections.emptyEnumeration());

        Pkcs11SignatureService service = new Pkcs11SignatureService(ks);
        ValidateRequest req = new ValidateRequest();
        req.setContent("documento");
        req.setSignature(Base64.getEncoder().encodeToString("assinatura-qualquer".getBytes()));

        SignatureResult result = service.validate(req);

        assertFalse(result.isValid());
        assertEquals("Assinatura é inválida", result.getMessage());
    }

    @Test
    void validate_encodingBase64Invalido_retornaErro() throws Exception {
        KeyStore ks = mock(KeyStore.class);

        Pkcs11SignatureService service = new Pkcs11SignatureService(ks);
        ValidateRequest req = new ValidateRequest();
        req.setContent("documento");
        req.setSignature("não é base64 válido!!!@@@");

        SignatureResult result = service.validate(req);

        assertFalse(result.isValid());
        assertTrue(result.getMessage().contains("encoding inválido"));
    }

    @Test
    void validate_excecaoNoKeyStore_retornaErro() throws Exception {
        KeyStore ks = mock(KeyStore.class);
        when(ks.aliases()).thenThrow(new RuntimeException("HSM offline"));

        Pkcs11SignatureService service = new Pkcs11SignatureService(ks);
        ValidateRequest req = new ValidateRequest();
        req.setContent("documento");
        req.setSignature(Base64.getEncoder().encodeToString("sig".getBytes()));

        SignatureResult result = service.validate(req);

        assertFalse(result.isValid());
        assertTrue(result.getMessage().contains("PKCS#11"));
    }
}
