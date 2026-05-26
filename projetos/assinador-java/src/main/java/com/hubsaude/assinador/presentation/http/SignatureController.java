package com.hubsaude.assinador.presentation.http;

import com.hubsaude.assinador.application.SignUseCase;
import com.hubsaude.assinador.application.ValidateUseCase;
import com.hubsaude.assinador.domain.model.SignRequest;
import com.hubsaude.assinador.domain.model.SignatureResult;
import com.hubsaude.assinador.domain.model.ValidateRequest;
import com.hubsaude.assinador.presentation.http.dto.SignHttpRequest;
import com.hubsaude.assinador.presentation.http.dto.SignatureHttpResponse;
import com.hubsaude.assinador.presentation.http.dto.ValidateHttpRequest;
import org.springframework.web.bind.annotation.PostMapping;
import org.springframework.web.bind.annotation.RequestBody;
import org.springframework.web.bind.annotation.RestController;

@RestController
public class SignatureController {

    private final SignUseCase signUseCase;
    private final ValidateUseCase validateUseCase;

    public SignatureController(SignUseCase signUseCase, ValidateUseCase validateUseCase) {
        this.signUseCase = signUseCase;
        this.validateUseCase = validateUseCase;
    }

    @PostMapping("/sign")
    public SignatureHttpResponse sign(@RequestBody SignHttpRequest request) {
        SignRequest domainRequest = new SignRequest();
        domainRequest.setContent(request.getContent());
        domainRequest.setToken(request.getToken());
        return toResponse(signUseCase.execute(domainRequest));
    }

    @PostMapping("/validate")
    public SignatureHttpResponse validate(@RequestBody ValidateHttpRequest request) {
        ValidateRequest domainRequest = new ValidateRequest();
        domainRequest.setContent(request.getContent());
        domainRequest.setSignature(request.getSignature());
        return toResponse(validateUseCase.execute(domainRequest));
    }

    private SignatureHttpResponse toResponse(SignatureResult result) {
        return new SignatureHttpResponse(result.getSignature(), result.isValid(), result.getMessage());
    }
}
