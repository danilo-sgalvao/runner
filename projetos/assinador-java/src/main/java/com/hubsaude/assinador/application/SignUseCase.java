package com.hubsaude.assinador.application;

import com.hubsaude.assinador.application.validation.RequestValidator;
import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.domain.service.SignatureService;

public class SignUseCase {

    private final SignatureService service;
    private final RequestValidator validator;

    public SignUseCase(SignatureService service, RequestValidator validator) {
        this.service = service;
        this.validator = validator;
    }

    public SignatureResult execute(SignRequest request) {
        validator.validateSign(request);
        return service.sign(request);
    }
}
