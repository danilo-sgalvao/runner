package com.hubsaude.assinador.presentation.http;

import com.hubsaude.assinador.application.validation.ValidationException;
import com.hubsaude.assinador.presentation.http.dto.SignatureHttpResponse;
import org.springframework.http.HttpStatus;
import org.springframework.web.bind.annotation.ExceptionHandler;
import org.springframework.web.bind.annotation.ResponseStatus;
import org.springframework.web.bind.annotation.RestControllerAdvice;

@RestControllerAdvice
public class GlobalExceptionHandler {

    @ExceptionHandler(ValidationException.class)
    @ResponseStatus(HttpStatus.BAD_REQUEST)
    public SignatureHttpResponse handleValidation(ValidationException ex) {
        return new SignatureHttpResponse(null, false, ex.getMessage());
    }

    @ExceptionHandler(Exception.class)
    @ResponseStatus(HttpStatus.INTERNAL_SERVER_ERROR)
    public SignatureHttpResponse handleGeneral(Exception ex) {
        return new SignatureHttpResponse(null, false, "Erro interno do servidor");
    }
}
