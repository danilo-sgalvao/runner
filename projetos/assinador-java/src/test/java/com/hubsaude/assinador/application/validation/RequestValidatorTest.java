package com.hubsaude.assinador.application.validation;

import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.ValidateRequest;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class RequestValidatorTest {

    private final RequestValidator validator = new RequestValidator();

    // ----------------------------------------------------------- validateSign

    @Test
    void validateSign_conteudoValido_naoLancaExcecao() {
        SignRequest request = new SignRequest();
        request.setContent("documento");
        assertDoesNotThrow(() -> validator.validateSign(request));
    }

    @Test
    void validateSign_requestNulo_lancaExcecaoComMensagemContent() {
        ValidationException ex = assertThrows(ValidationException.class,
            () -> validator.validateSign(null));
        assertTrue(ex.getMessage().contains("content"));
    }

    @Test
    void validateSign_conteudoNulo_lancaExcecao() {
        SignRequest request = new SignRequest();
        ValidationException ex = assertThrows(ValidationException.class,
            () -> validator.validateSign(request));
        assertEquals("Parâmetro 'content' inválido ou ausente", ex.getMessage());
    }

    @Test
    void validateSign_conteudoVazio_lancaExcecao() {
        SignRequest request = new SignRequest();
        request.setContent("");
        assertThrows(ValidationException.class, () -> validator.validateSign(request));
    }

    @Test
    void validateSign_conteudoApenasEspacos_lancaExcecao() {
        SignRequest request = new SignRequest();
        request.setContent("   ");
        assertThrows(ValidationException.class, () -> validator.validateSign(request));
    }

    // ------------------------------------------------------- validateValidate

    @Test
    void validateValidate_parametrosValidos_naoLancaExcecao() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature("SIG==");
        assertDoesNotThrow(() -> validator.validateValidate(request));
    }

    @Test
    void validateValidate_conteudoNulo_lancaExcecaoComMensagemContent() {
        ValidateRequest request = new ValidateRequest();
        request.setSignature("SIG==");
        ValidationException ex = assertThrows(ValidationException.class,
            () -> validator.validateValidate(request));
        assertEquals("Parâmetro 'content' inválido ou ausente", ex.getMessage());
    }

    @Test
    void validateValidate_conteudoApenasEspacos_lancaExcecao() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("   ");
        request.setSignature("SIG==");
        assertThrows(ValidationException.class, () -> validator.validateValidate(request));
    }

    @Test
    void validateValidate_assinaturaNula_lancaExcecaoComMensagemSignature() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        ValidationException ex = assertThrows(ValidationException.class,
            () -> validator.validateValidate(request));
        assertEquals("Parâmetro 'signature' inválido ou ausente", ex.getMessage());
    }

    @Test
    void validateValidate_assinaturaVazia_lancaExcecao() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature("");
        assertThrows(ValidationException.class, () -> validator.validateValidate(request));
    }
}
