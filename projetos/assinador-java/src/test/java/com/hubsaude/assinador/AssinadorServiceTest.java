package com.hubsaude.assinador;

import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class AssinadorServiceTest {

    // ------------------------------------------------------------------ sign

    @Test
    void sign_conteudoValido_retornaAssinaturaComPrefixoCorreto() {
        String resultado = AssinadorService.sign("documento teste", "SHA256withRSA");
        assertTrue(resultado.startsWith("ASSINATURA-SIMULADA-SHA256withRSA-"),
            "Assinatura deveria começar com o prefixo correto");
    }

    @Test
    void sign_algoritmoSHA512_retornaAssinaturaComAlgoritmoCorreto() {
        String resultado = AssinadorService.sign("documento", "SHA512withRSA");
        assertTrue(resultado.startsWith("ASSINATURA-SIMULADA-SHA512withRSA-"),
            "Assinatura deveria conter o algoritmo SHA512withRSA");
    }

    @Test
    void sign_mesmosParametros_retornamMesmaAssinatura() {
        String a1 = AssinadorService.sign("conteudo fixo", "SHA256withRSA");
        String a2 = AssinadorService.sign("conteudo fixo", "SHA256withRSA");
        assertEquals(a1, a2, "Assinatura deveria ser determinística para os mesmos parâmetros");
    }

    @Test
    void sign_conteudosDistintos_retornamAssinaturasDiferentes() {
        String a1 = AssinadorService.sign("conteudo A", "SHA256withRSA");
        String a2 = AssinadorService.sign("conteudo B", "SHA256withRSA");
        assertNotEquals(a1, a2, "Conteúdos diferentes deveriam gerar assinaturas diferentes");
    }

    @Test
    void sign_conteudoNulo_lancaIllegalArgumentException() {
        IllegalArgumentException ex = assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.sign(null, "SHA256withRSA"));
        assertTrue(ex.getMessage().contains("--content"),
            "Mensagem de erro deveria mencionar --content");
    }

    @Test
    void sign_conteudoVazio_lancaIllegalArgumentException() {
        assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.sign("", "SHA256withRSA"));
    }

    @Test
    void sign_conteudoApenasEspacos_lancaIllegalArgumentException() {
        assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.sign("   ", "SHA256withRSA"));
    }

    @Test
    void sign_algoritmoInvalido_lancaIllegalArgumentException() {
        IllegalArgumentException ex = assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.sign("conteudo", "MD5withRSA"));
        assertTrue(ex.getMessage().contains("MD5withRSA"),
            "Mensagem de erro deveria mencionar o algoritmo inválido fornecido");
    }

    @Test
    void sign_algoritmoNulo_lancaIllegalArgumentException() {
        assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.sign("conteudo", null));
    }

    @Test
    void sign_algoritmoVazio_lancaIllegalArgumentException() {
        assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.sign("conteudo", ""));
    }

    // --------------------------------------------------------------- validate

    @Test
    void validate_assinaturaSimulada_retornaTrue() {
        boolean resultado = AssinadorService.validate(
            "qualquer conteudo",
            "ASSINATURA-SIMULADA-SHA256withRSA-ABC123"
        );
        assertTrue(resultado, "Assinatura com prefixo correto deveria ser válida");
    }

    @Test
    void validate_assinaturaArbitraria_retornaFalse() {
        boolean resultado = AssinadorService.validate("conteudo", "assinatura-invalida");
        assertFalse(resultado, "Assinatura sem prefixo não deveria ser válida");
    }

    @Test
    void validate_assinaturaVaziaString_lancaIllegalArgumentException() {
        assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.validate("conteudo", ""));
    }

    @Test
    void validate_assinaturaNula_lancaIllegalArgumentException() {
        IllegalArgumentException ex = assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.validate("conteudo", null));
        assertTrue(ex.getMessage().contains("--signature"),
            "Mensagem de erro deveria mencionar --signature");
    }

    @Test
    void validate_conteudoNulo_lancaIllegalArgumentException() {
        assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.validate(null, "ASSINATURA-SIMULADA-SHA256withRSA-X"));
    }

    @Test
    void validate_conteudoVazio_lancaIllegalArgumentException() {
        assertThrows(IllegalArgumentException.class,
            () -> AssinadorService.validate("", "ASSINATURA-SIMULADA-SHA256withRSA-X"));
    }

    @Test
    void validate_fluxoCompleto_signEntaoValidate_retornaTrue() {
        String conteudo   = "documento importante";
        String assinatura = AssinadorService.sign(conteudo, "SHA256withRSA");
        boolean valida    = AssinadorService.validate(conteudo, assinatura);
        assertTrue(valida, "Assinatura gerada pelo sign deveria ser válida no validate");
    }
}
