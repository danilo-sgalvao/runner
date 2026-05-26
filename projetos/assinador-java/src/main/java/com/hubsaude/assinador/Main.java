package com.hubsaude.assinador;

import com.hubsaude.assinador.application.SignUseCase;
import com.hubsaude.assinador.application.ValidateUseCase;
import com.hubsaude.assinador.application.validation.RequestValidator;
import com.hubsaude.assinador.application.validation.ValidationException;
import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.ValidateRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.domain.service.FakeSignatureService;
import com.hubsaude.assinador.domain.service.SignatureService;
import com.hubsaude.assinador.infrastructure.json.JsonMapper;

public class Main {

    private static final SignatureService service  = new FakeSignatureService();
    private static final RequestValidator validator = new RequestValidator();
    private static final SignUseCase      signUseCase     = new SignUseCase(service, validator);
    private static final ValidateUseCase  validateUseCase = new ValidateUseCase(service, validator);

    public static void main(String[] args) {
        if (args.length == 0) {
            System.err.println("Erro: nenhum comando fornecido.");
            System.err.println("Uso: assinador <comando> [opções]");
            System.err.println("Comandos disponíveis: sign, validate");
            System.exit(1);
        }

        switch (args[0]) {
            case "sign"     -> handleSign(args);
            case "validate" -> handleValidate(args);
            default -> {
                System.err.println("Erro: comando desconhecido: " + args[0]);
                System.err.println("Comandos disponíveis: sign, validate");
                System.exit(1);
            }
        }
    }

    private static void handleSign(String[] args) {
        SignRequest request = new SignRequest();

        for (int i = 1; i < args.length; i++) {
            if ("--content".equals(args[i]) && i + 1 < args.length) {
                request.setContent(args[++i]);
            }
        }

        try {
            printResponse(signUseCase.execute(request));
        } catch (ValidationException e) {
            printResponse(new SignatureResult(null, false, e.getMessage()));
        }
    }

    private static void handleValidate(String[] args) {
        ValidateRequest request = new ValidateRequest();

        for (int i = 1; i < args.length; i++) {
            switch (args[i]) {
                case "--content"   -> { if (i + 1 < args.length) request.setContent(args[++i]); }
                case "--signature" -> { if (i + 1 < args.length) request.setSignature(args[++i]); }
            }
        }

        try {
            printResponse(validateUseCase.execute(request));
        } catch (ValidationException e) {
            printResponse(new SignatureResult(null, false, e.getMessage()));
        }
    }

    private static void printResponse(SignatureResult response) {
        String json = JsonMapper.toJson(response);
        if (response.isValid()) {
            System.out.println(json);
        } else {
            System.err.println(json);
            System.exit(1);
        }
    }
}
