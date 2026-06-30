// RecipeApp JavaScript utilities

// Modal functions
function showLoginModal() {
    document.getElementById('loginModal').style.display = 'flex';
}

function hideLoginModal() {
    document.getElementById('loginModal').style.display = 'none';
}

function showRegisterModal() {
    document.getElementById('registerModal').style.display = 'flex';
}

function hideRegisterModal() {
    document.getElementById('registerModal').style.display = 'none';
}

// Ingredient management for recipe form
let ingredientCount = 1;

function addIngredient() {
    const ingredientsList = document.getElementById('ingredients-list');
    const newIngredient = document.createElement('div');
    newIngredient.className = 'ingredient-row';
    newIngredient.innerHTML = `
        <input type="text" name="ingredients[${ingredientCount}].name" placeholder="ej. Arroz" required class="form-input">
        <input type="text" name="ingredients[${ingredientCount}].quantity" placeholder="400" required class="form-input">
        <input type="text" name="ingredients[${ingredientCount}].unit" placeholder="g" required class="form-input">
        <button type="button" onclick="removeIngredient(this)" class="btn-remove">✕</button>
    `;
    ingredientsList.appendChild(newIngredient);
    ingredientCount++;
}

function removeIngredient(button) {
    const ingredientItem = button.closest('.ingredient-row');
    ingredientItem.remove();
}

// Instruction management for recipe form
let instructionCount = 1;

function addInstruction() {
    const instructionsList = document.getElementById('instructions-list');
    const newInstruction = document.createElement('div');
    newInstruction.className = 'instruction-row';
    newInstruction.innerHTML = `
        <span class="step-label">${instructionCount + 1}.</span>
        <textarea name="instructions[${instructionCount}]" placeholder="Paso ${instructionCount + 1}" rows="2" required class="form-textarea" style="flex:1;"></textarea>
        <button type="button" onclick="removeInstruction(this)" class="btn-remove" style="align-self:flex-start; margin-top:0.4rem;">✕</button>
    `;
    instructionsList.appendChild(newInstruction);
    instructionCount++;
}

function removeInstruction(button) {
    const instructionItem = button.closest('.instruction-row');
    instructionItem.remove();
    renumberInstructionRows();
}

function renumberInstructionRows() {
    const instructions = document.querySelectorAll('.instruction-row');

    instructions.forEach((item, index) => {
        const label = item.querySelector('.step-label');
        const textarea = item.querySelector('textarea');

        if (label) {
            label.textContent = `${index + 1}.`;
        }

        if (textarea) {
            textarea.name = `instructions[${index}]`;
            textarea.placeholder = `Paso ${index + 1}`;
        }
    });

    instructionCount = instructions.length;
}

// Mobile menu toggle
function toggleMobileMenu() {
    const nav = document.querySelector('.site-nav');
    const button = document.getElementById('mobile-menu-btn');

    if (!nav) {
        return;
    }

    const isOpen = nav.classList.toggle('open');

    if (button) {
        button.setAttribute('aria-expanded', String(isOpen));
    }
}

// Close modals when clicking outside
document.addEventListener('click', function(event) {
    const loginModal = document.getElementById('loginModal');
    const registerModal = document.getElementById('registerModal');
    
    if (event.target === loginModal) {
        hideLoginModal();
    }
    if (event.target === registerModal) {
        hideRegisterModal();
    }
});

// HTMX event handlers
document.addEventListener('htmx:afterRequest', function(event) {
    const target = event.target;
    
    // Handle successful authentication
    if (target.id === 'loginModal' || target.id === 'registerModal') {
        if (event.detail.successful) {
            // Redirect to reload the page with new auth state
            window.location.reload();
        }
    }
    
    // Handle recipe creation
    if (target.id === 'form-message') {
        if (event.detail.successful) {
            // Redirect to recipes page on successful creation
            setTimeout(() => {
                window.location.href = '/recipes';
            }, 1500);
        }
    }
});

// Form validation helpers
function validateForm(formData) {
    const errors = [];
    
    // Basic validation
    if (!formData.get('title')?.trim()) {
        errors.push('Recipe title is required');
    }
    if (!formData.get('description')?.trim()) {
        errors.push('Description is required');
    }
    if (!formData.get('difficulty')) {
        errors.push('Difficulty level is required');
    }
    
    return errors;
}

// Show flash messages
function showFlashMessage(message, type = 'info') {
    const flashDiv = document.createElement('div');
    flashDiv.className = `fixed top-20 right-4 p-4 rounded-lg shadow-lg z-50 ${
        type === 'error' ? 'bg-red-500 text-white' :
        type === 'success' ? 'bg-green-500 text-white' :
        type === 'warning' ? 'bg-yellow-500 text-white' :
        'bg-blue-500 text-white'
    }`;
    flashDiv.textContent = message;
    
    document.body.appendChild(flashDiv);
    
    setTimeout(() => {
        flashDiv.remove();
    }, 5000);
}