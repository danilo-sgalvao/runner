package com.hubsaude.assinador.application.validation;

import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.ValidateRequest;

public class RequestValidator {

    public void validateSign(SignRequest request) {
        if (request == null || isBlank(request.getContent())) {
            throw new ValidationException("Parâmetro 'content' inválido ou ausente");
        }
    }

    public void validateValidate(ValidateRequest request) {
        if (request == null || isBlank(request.getContent())) {
            throw new ValidationException("Parâmetro 'content' inválido ou ausente");
        }
        if (isBlank(request.getSignature())) {
            throw new ValidationException("Parâmetro 'signature' inválido ou ausente");
        }
    }

    private boolean isBlank(String s) {
        return s == null || s.isBlank();
    }
}
