package com.hubsaude.assinador.infrastructure.config;

import com.hubsaude.assinador.application.SignUseCase;
import com.hubsaude.assinador.application.ValidateUseCase;
import com.hubsaude.assinador.application.validation.RequestValidator;
import com.hubsaude.assinador.domain.service.FakeSignatureService;
import com.hubsaude.assinador.domain.service.SignatureService;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class AppConfig {

    @Bean
    public SignatureService signatureService() {
        return new FakeSignatureService();
    }

    @Bean
    public RequestValidator requestValidator() {
        return new RequestValidator();
    }

    @Bean
    public SignUseCase signUseCase(SignatureService signatureService, RequestValidator requestValidator) {
        return new SignUseCase(signatureService, requestValidator);
    }

    @Bean
    public ValidateUseCase validateUseCase(SignatureService signatureService, RequestValidator requestValidator) {
        return new ValidateUseCase(signatureService, requestValidator);
    }
}
