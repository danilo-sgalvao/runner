package com.hubsaude.assinador.application;

import com.hubsaude.assinador.application.validation.RequestValidator;
import com.hubsaude.assinador.application.validation.ValidationException;
import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.ValidateRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.domain.service.FakeSignatureService;
import com.hubsaude.assinador.domain.service.SignatureService;
import org.junit.jupiter.api.Test;

import static org.junit.jupiter.api.Assertions.*;

class UseCasesTest {

    private final SignatureService service  = new FakeSignatureService();
    private final RequestValidator validator = new RequestValidator();
    private final SignUseCase      signUseCase     = new SignUseCase(service, validator);
    private final ValidateUseCase  validateUseCase = new ValidateUseCase(service, validator);

    // ---------------------------------------------------------- SignUseCase

    @Test
    void sign_entradaValida_retornaAssinatura() {
        SignRequest request = new SignRequest();
        request.setContent("documento");

        SignatureResult result = signUseCase.execute(request);

        assertTrue(result.isValid());
        assertEquals(FakeSignatureService.FAKE_SIGNATURE, result.getSignature());
    }

    @Test
    void sign_contentVazio_propagaValidationException() {
        SignRequest request = new SignRequest();
        request.setContent("");

        assertThrows(ValidationException.class, () -> signUseCase.execute(request));
    }

    // ------------------------------------------------------- ValidateUseCase

    @Test
    void validate_assinaturaCorreta_retornaValida() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature(FakeSignatureService.FAKE_SIGNATURE);

        SignatureResult result = validateUseCase.execute(request);

        assertTrue(result.isValid());
    }

    @Test
    void validate_signatureVazia_propagaValidationException() {
        ValidateRequest request = new ValidateRequest();
        request.setContent("documento");
        request.setSignature("");

        assertThrows(ValidationException.class, () -> validateUseCase.execute(request));
    }

    // -------------------------------------------------------- fluxo completo

    @Test
    void fluxoCompleto_signEntaoValidate_retornaValida() {
        SignRequest signReq = new SignRequest();
        signReq.setContent("documento importante");
        SignatureResult signed = signUseCase.execute(signReq);
        assertTrue(signed.isValid());

        ValidateRequest validateReq = new ValidateRequest();
        validateReq.setContent("documento importante");
        validateReq.setSignature(signed.getSignature());
        SignatureResult validated = validateUseCase.execute(validateReq);

        assertTrue(validated.isValid());
    }
}
