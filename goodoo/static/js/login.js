document.addEventListener('DOMContentLoaded', function() {
    const loginForm = document.getElementById('loginForm');
    const errorMessage = document.getElementById('errorMessage');
    const successMessage = document.getElementById('successMessage');

    loginForm.addEventListener('submit', async function(e) {
        e.preventDefault();
        
        const formData = new FormData(loginForm);
        const loginData = {
            login: formData.get('login'),
            password: formData.get('password')
        };

        try {
            const response = await fetch('/auth/login', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify(loginData),
                credentials: 'include'
            });

            const result = await response.json();

            if (response.ok) {
                showSuccess('Login successful! Redirecting...');
                setTimeout(() => {
                    window.location.href = '/';
                }, 1500);
            } else {
                showError(result.error || 'Login failed. Please try again.');
            }
        } catch (error) {
            showError('Network error. Please check your connection and try again.');
        }
    });

    function showError(message) {
        errorMessage.textContent = message;
        errorMessage.style.display = 'block';
        successMessage.style.display = 'none';
    }

    function showSuccess(message) {
        successMessage.textContent = message;
        successMessage.style.display = 'block';
        errorMessage.style.display = 'none';
    }
});