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
        SignatureService service         = new FakeSignatureService();
        RequestValidator validator       = new RequestValidator();
        SignUseCase      signUseCase     = new SignUseCase(service, validator);
        ValidateUseCase  validateUseCase = new ValidateUseCase(service, validator);
        CliPresenter     presenter       = new CliPresenter();
        CliRunner        runner          = new CliRunner(signUseCase, validateUseCase, presenter);

        runner.run(args);
    }
}
