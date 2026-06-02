package com.hubsaude.assinador.infrastructure.config;

import com.hubsaude.assinador.application.SignUseCase;
import com.hubsaude.assinador.application.ValidateUseCase;
import com.hubsaude.assinador.application.validation.RequestValidator;
import com.hubsaude.assinador.domain.service.FakeSignatureService;
import com.hubsaude.assinador.domain.service.SignatureService;
import com.hubsaude.assinador.infrastructure.pkcs11.Pkcs11Config;
import com.hubsaude.assinador.infrastructure.pkcs11.Pkcs11ServiceFactory;
import org.springframework.context.annotation.Bean;
import org.springframework.context.annotation.Configuration;

@Configuration
public class AppConfig {

    @Bean
    public SignatureService signatureService() {
        Pkcs11Config pkcs11Config = Pkcs11Config.fromEnvironment();
        if (pkcs11Config != null) {
            try {
                return Pkcs11ServiceFactory.create(pkcs11Config);
            } catch (Exception e) {
                System.err.println("Aviso: PKCS#11 configurado mas não disponível: " + e.getMessage());
                System.err.println("Usando serviço simulado (fake).");
            }
        }
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
