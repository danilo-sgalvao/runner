package com.hubsaude.assinador;

import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.ValidateRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.domain.service.FakeSignatureService;
import com.hubsaude.assinador.domain.service.SignatureService;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class FakeSignatureServiceTest {

    private final SignatureService service = new FakeSignatureService();

    // ------------------------------------------------------------------ sign

    @Test
    void sign_conteudoValido_retornaAssinaturaSimulada() {
        SignRequest request = new SignRequest();
        request.setContent("documento teste");

        SignatureResult result = service.sign(request);

        assertNotNull(result);
        assertTrue(result.isValid());
        assertEquals(FakeSignatureService.FAKE_SIGNATURE, result.getSignature());
        assertEquals("Assinatura criada com sucesso", result.getMessage());
    }

    // --------------------------------------------------------------- validate

    @Test
    void validate_assinaturaCorreta_retornaValida() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature(FakeSignatureService.FAKE_SIGNATURE);

        SignatureResult result = service.validate(request);

        assertNotNull(result);
        assertTrue(result.isValid());
        assertEquals("Assinatura é válida", result.getMessage());
    }

    @Test
    void validate_assinaturaErrada_retornaInvalida() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature("ASSINATURA-ERRADA");

        SignatureResult result = service.validate(request);

        assertFalse(result.isValid());
        assertEquals("Assinatura é inválida", result.getMessage());
    }

    // --------------------------------------------------------- fluxo completo

    @Test
    void fluxoCompleto_signEntaoValidate_retornaValida() {
        SignRequest signReq = new SignRequest();
        signReq.setContent("documento importante");
        SignatureResult signed = service.sign(signReq);
        assertTrue(signed.isValid());

        ValidateRequest validateReq = new ValidateRequest();
        validateReq.setContent("documento importante");
        validateReq.setSignature(signed.getSignature());
        SignatureResult validated = service.validate(validateReq);

        assertTrue(validated.isValid());
    }
}
