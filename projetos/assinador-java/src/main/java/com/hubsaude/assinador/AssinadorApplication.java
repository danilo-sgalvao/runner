package com.hubsaude.assinador;

import com.hubsaude.assinador.application.SignUseCase;
import com.hubsaude.assinador.application.ValidateUseCase;
import com.hubsaude.assinador.application.validation.RequestValidator;
import com.hubsaude.assinador.domain.service.FakeSignatureService;
import com.hubsaude.assinador.domain.service.SignatureService;
import com.hubsaude.assinador.presentation.cli.CliPresenter;
import com.hubsaude.assinador.presentation.cli.CliRunner;

public class AssinadorApplication {

    public static void main(String[] args) {
        if (args.length > 0 && "serve".equals(args[0])) {
            // Ponto de extensão da Sprint 3 (US-02.4): modo servidor HTTP via Spring Boot.
            System.err.println("Modo servidor (serve) ainda não implementado.");
            System.exit(1);
            return;
        }

        SignatureService service         = new FakeSignatureService();
        RequestValidator validator       = new RequestValidator();
        SignUseCase      signUseCase     = new SignUseCase(service, validator);
        ValidateUseCase  validateUseCase = new ValidateUseCase(service, validator);
        CliPresenter     presenter       = new CliPresenter();
        CliRunner        runner          = new CliRunner(signUseCase, validateUseCase, presenter);

        runner.run(args);
    }
}
