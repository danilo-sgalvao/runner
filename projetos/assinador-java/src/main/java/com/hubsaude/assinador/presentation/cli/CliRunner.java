package com.hubsaude.assinador.presentation.cli;

import com.hubsaude.assinador.application.SignUseCase;
import com.hubsaude.assinador.application.ValidateUseCase;
import com.hubsaude.assinador.application.validation.ValidationException;
import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.domain.model.ValidateRequest;

public class CliRunner {

    private final SignUseCase     signUseCase;
    private final ValidateUseCase validateUseCase;
    private final CliPresenter    presenter;

    public CliRunner(SignUseCase signUseCase, ValidateUseCase validateUseCase, CliPresenter presenter) {
        this.signUseCase     = signUseCase;
        this.validateUseCase = validateUseCase;
        this.presenter       = presenter;
    }

    public void run(String[] args) {
        if (args.length == 0) {
            presenter.presentError("Erro: nenhum comando fornecido.");
            presenter.presentError("Uso: assinador <comando> [opções]");
            presenter.presentError("Comandos disponíveis: sign, validate");
            System.exit(1);
            return;
        }

        switch (args[0]) {
            case "sign"     -> handleSign(args);
            case "validate" -> handleValidate(args);
            default -> {
                presenter.presentError("Erro: comando desconhecido: " + args[0]);
                presenter.presentError("Comandos disponíveis: sign, validate");
                System.exit(1);
            }
        }
    }

    private void handleSign(String[] args) {
        SignRequest request = new SignRequest();

        for (int i = 1; i < args.length; i++) {
            if ("--content".equals(args[i]) && i + 1 < args.length) {
                request.setContent(args[++i]);
            }
        }

        try {
            presenter.present(signUseCase.execute(request));
        } catch (ValidationException e) {
            presenter.present(new SignatureResult(null, false, e.getMessage()));
        }
    }

    private void handleValidate(String[] args) {
        ValidateRequest request = new ValidateRequest();

        for (int i = 1; i < args.length; i++) {
            switch (args[i]) {
                case "--content"   -> { if (i + 1 < args.length) request.setContent(args[++i]); }
                case "--signature" -> { if (i + 1 < args.length) request.setSignature(args[++i]); }
            }
        }

        try {
            presenter.present(validateUseCase.execute(request));
        } catch (ValidationException e) {
            presenter.present(new SignatureResult(null, false, e.getMessage()));
        }
    }
}
