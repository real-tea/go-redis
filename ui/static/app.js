// --- DOM Elements ---
const authContainer = document.getElementById('auth-container');
const appContainer = document.getElementById('app-container');
const authForm = document.getElementById('auth-form');
const authTitle = document.getElementById('auth-title');
const toggleAuthLink = document.getElementById('toggle-auth');
const logoutBtn = document.getElementById('logout-btn');

const balance = document.getElementById('balance');
const transactionList = document.getElementById('transaction-list');
const transactionForm = document.getElementById('transaction-form');
const descriptionInput = document.getElementById('description');
const amountInput = document.getElementById('amount');

// --- State ---
let isLoginMode = true;
let credentials = null; // To store 'username:password' for basic auth

// --- API URLs ---
const REGISTER_URL = '/api/register';
const LOGIN_URL = '/api/login';
const TRANSACTIONS_URL = '/api/transactions';

// --- Functions ---

// Function to get Auth Header
function getAuthHeader() {
    if (!credentials) return {};
    return { 'Authorization': 'Basic ' + btoa(credentials) }; // btoa creates base64 encoding
}

// Show/hide main app vs auth form
function showApp(isLoggedIn) {
    if (isLoggedIn) {
        authContainer.style.display = 'none';
        appContainer.style.display = 'block';
        getTransactions();
    } else {
        authContainer.style.display = 'block';
        appContainer.style.display = 'none';
    }
}

// Fetch transactions from the backend
async function getTransactions() {
    try {
        const res = await fetch(TRANSACTIONS_URL, { headers: getAuthHeader() });
        
        // This 'if' block is the fix. It checks if the request was successful.
        if (!res.ok) {
            // If token is invalid (401 Unauthorized), log the user out.
            if(res.status === 401) {
                console.error("Authentication failed. Logging out.");
                handleLogout();
            }
            throw new Error('Failed to fetch transactions');
        }

        const data = await res.json(); // This line is now only run on success
        
        transactionList.innerHTML = '';
        if (data && Array.isArray(data)) {
            data.forEach(addTransactionDOM);
        }
        updateBalance();
    } catch (error) {
        console.error('Error in getTransactions:', error);
    }
}


// Add a new transaction via the backend
async function addTransaction(e) {
    e.preventDefault();
    if (descriptionInput.value.trim() === '' || amountInput.value.trim() === '') {
        alert('Please add a description and amount');
        return;
    }

    const newTransaction = {
        description: descriptionInput.value,
        amount: +amountInput.value,
    };

    try {
        const res = await fetch(TRANSACTIONS_URL, {
            method: 'POST',
            headers: { ...getAuthHeader(), 'Content-Type': 'application/json' },
            body: JSON.stringify(newTransaction),
        });
        
        if (!res.ok) throw new Error('Failed to add transaction');

        const data = await res.json();
        addTransactionDOM(data);
        updateBalance();
        transactionForm.reset();
    } catch (error) {
        console.error('Error:', error);
    }
}

// Add transaction to the DOM list
function addTransactionDOM(transaction) {
    const sign = transaction.amount < 0 ? '-' : '+';
    const item = document.createElement('tr');
    const amountClass = transaction.amount < 0 ? 'expense' : 'income';
    item.innerHTML = `
        <td>${transaction.description}</td>
        <td class="${amountClass}">${sign}₹${Math.abs(transaction.amount).toFixed(2)}</td>
        <td>${new Date(transaction.date).toLocaleDateString()}</td>
    `;
    transactionList.appendChild(item);
}

// Update the balance total
function updateBalance() {
    const rows = transactionList.querySelectorAll('tr');
    let total = 0;
    rows.forEach(row => {
        const amountCell = row.querySelector('td:nth-child(2)');
        if (amountCell) {
            const amountValue = parseFloat(amountCell.textContent.replace(/[₹+-]/g, ''));
            total += amountCell.classList.contains('expense') ? -amountValue : amountValue;
        }
    });
    balance.innerText = `₹${total.toFixed(2)}`;
}

// Handle Auth Form Submission (Login/Register)
async function handleAuth(e) {
    e.preventDefault();
    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    if (!username || !password) {
        alert('Please enter username and password');
        return;
    }
    
    credentials = `${username}:${password}`;

    if (isLoginMode) {
        // --- Login ---
        try {
            const res = await fetch(LOGIN_URL, { headers: getAuthHeader() });
            if (!res.ok) throw new Error('Invalid credentials');
            alert('Login successful!');
            localStorage.setItem('moneyTrackerCredentials', credentials);
            showApp(true);
        } catch (error) {
            alert('Login failed. Please check your username and password.');
            credentials = null;
        }
    } else {
        // --- Register ---
        try {
            const res = await fetch(REGISTER_URL, {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify({ username, password })
            });
            if (!res.ok) throw new Error('Registration failed');
            alert('Registration successful! Please log in.');
            toggleAuthMode(); // Switch to login view
            authForm.reset();
        } catch (error) {
            alert('Registration failed. The username may already be taken.');
        }
        credentials = null;
    }
}

// Handle Logout
function handleLogout() {
    credentials = null;
    localStorage.removeItem('moneyTrackerCredentials');
    showApp(false);
}

// Toggle between Login and Register modes
function toggleAuthMode(e) {
    if(e) e.preventDefault();
    isLoginMode = !isLoginMode;
    authTitle.innerText = isLoginMode ? 'Login' : 'Register';
    toggleAuthLink.innerHTML = isLoginMode ? 'Don\'t have an account? <a href="#">Register</a>' : 'Already have an account? <a href="#">Login</a>';
    authForm.reset();
}

// --- Event Listeners ---
authForm.addEventListener('submit', handleAuth);
transactionForm.addEventListener('submit', addTransaction);
toggleAuthLink.addEventListener('click', toggleAuthMode);
logoutBtn.addEventListener('click', handleLogout);

// --- Initial Load ---
// Check if user was previously logged in
const savedCreds = localStorage.getItem('moneyTrackerCredentials');
if (savedCreds) {
    credentials = savedCreds;
    // Let's verify the credentials before showing the app
    fetch(TRANSACTIONS_URL, { headers: { 'Authorization': 'Basic ' + btoa(savedCreds) } })
        .then(res => {
            if (res.ok) {
                showApp(true);
            } else {
                handleLogout();
            }
        });
} else {
    showApp(false);
}
