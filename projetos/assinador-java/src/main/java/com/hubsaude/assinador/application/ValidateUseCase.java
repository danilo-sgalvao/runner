package com.hubsaude.assinador.application;

import com.hubsaude.assinador.application.validation.RequestValidator;
import com.hubsaude.assinador.domain.model.ValidateRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.domain.service.SignatureService;

public class ValidateUseCase {

    private final SignatureService service;
    private final RequestValidator validator;

    public ValidateUseCase(SignatureService service, RequestValidator validator) {
        this.service = service;
        this.validator = validator;
    }

    public SignatureResult execute(ValidateRequest request) {
        validator.validateValidate(request);
        return service.validate(request);
    }
}
