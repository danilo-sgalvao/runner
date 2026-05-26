package com.hubsaude.assinador.domain.service;

import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.ValidateRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;

public class FakeSignatureService implements SignatureService {

    public static final String FAKE_SIGNATURE = "MOCKED_SIGNATURE_BASE64_==";

    @Override
    public SignatureResult sign(SignRequest request) {
        return new SignatureResult(FAKE_SIGNATURE, true, "Assinatura criada com sucesso");
    }

    @Override
    public SignatureResult validate(ValidateRequest request) {
        boolean isValid = FAKE_SIGNATURE.equals(request.getSignature());
        return new SignatureResult(
            request.getSignature(),
            isValid,
            isValid ? "Assinatura é válida" : "Assinatura é inválida"
        );
    }
}
