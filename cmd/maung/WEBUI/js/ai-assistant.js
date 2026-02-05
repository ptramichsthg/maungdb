// ===========================
// AI ASSISTANT WIDGET (PROFESSIONAL EDITION v2)
// ===========================

class AIAssistant {
  constructor() {
    this.isOpen = false;
    this.chatHistory = [];
    this.isTyping = false;

    // Bind methods
    this.toggle = this.toggle.bind(this);
    this.close = this.close.bind(this);
    this.sendMessage = this.sendMessage.bind(this);
    this.clearHistory = this.clearHistory.bind(this);
    this.toggleMenu = this.toggleMenu.bind(this);

    this.init();
    this.loadHistory();
  }

  init() {
    this.createWidget();

    // Event listeners
    document.getElementById('ai-toggle-btn').addEventListener('click', this.toggle);
    document.getElementById('ai-close-btn').addEventListener('click', this.close);
    document.getElementById('ai-send-btn').addEventListener('click', this.sendMessage);
    document.getElementById('ai-menu-btn').addEventListener('click', this.toggleMenu);
    document.getElementById('ai-new-chat-btn').addEventListener('click', this.clearHistory);

    // Close menu when clicking outside
    document.addEventListener('click', (e) => {
      const menu = document.getElementById('ai-menu-dropdown');
      const btn = document.getElementById('ai-menu-btn');
      if (!menu.classList.contains('hidden') && !menu.contains(e.target) && !btn.contains(e.target)) {
        menu.classList.add('hidden');
      }
    });

    const input = document.getElementById('ai-input');
    input.addEventListener('keydown', (e) => {
      if (e.key === 'Enter' && !e.shiftKey) {
        e.preventDefault();
        this.sendMessage();
      }
    });

    // Auto-resize
    input.addEventListener('input', () => {
      input.style.height = 'auto';
      input.style.height = Math.min(input.scrollHeight, 120) + 'px';
    });

    // Initialize Icons
    this.refreshIcons();

    this.showWelcome();
  }

  refreshIcons() {
    if (window.lucide) window.lucide.createIcons();
  }

  createWidget() {
    const widgetHTML = `
      <!-- Floating Button -->
      <button id="ai-toggle-btn" 
        class="fixed bottom-6 right-6 z-[9999] w-14 h-14 bg-slate-900 hover:bg-slate-800 text-white rounded-full shadow-xl hover:shadow-2xl hover:scale-105 transition-all duration-300 flex items-center justify-center group overflow-hidden border border-slate-700">
        <i data-lucide="bot" class="w-7 h-7 transition-transform duration-300 group-hover:rotate-12"></i>
        <div class="absolute inset-0 bg-white/10 rounded-full scale-0 group-hover:scale-150 transition-transform duration-500 opacity-0 group-hover:opacity-100"></div>
      </button>

      <!-- Chat Container -->
      <div id="ai-chat-container" 
        class="fixed bottom-24 right-6 w-[400px] h-[600px] bg-white dark:bg-slate-900 rounded-2xl shadow-2xl border border-slate-200 dark:border-slate-800 flex flex-col overflow-hidden transition-all duration-300 origin-bottom-right transform scale-95 opacity-0 pointer-events-none z-[9998]">
        
        <!-- Header -->
        <div class="bg-slate-50 dark:bg-slate-950 px-5 py-4 border-b border-slate-200 dark:border-slate-800 flex justify-between items-center shrink-0 relative">
          <div class="flex items-center gap-3">
            <div class="w-8 h-8 rounded-lg bg-indigo-600 flex items-center justify-center text-white shadow-md shadow-indigo-500/20">
              <i data-lucide="sparkles" class="w-4 h-4"></i>
            </div>
            <div>
              <h3 class="font-bold text-slate-800 dark:text-white text-sm leading-tight">MaungDB Assistant</h3>
              <div class="flex items-center gap-1.5 mt-0.5">
                <span class="w-1.5 h-1.5 rounded-full bg-emerald-500"></span>
                <span class="text-xs text-slate-500 dark:text-slate-400 font-medium">Online</span>
              </div>
            </div>
          </div>
          
          <div class="flex items-center gap-1">
              <!-- Menu Button -->
              <div class="relative">
                  <button id="ai-menu-btn" class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors">
                    <i data-lucide="more-vertical" class="w-5 h-5"></i>
                  </button>
                  
                  <!-- Dropdown Menu -->
                  <div id="ai-menu-dropdown" class="hidden absolute right-0 top-full mt-2 w-48 bg-white dark:bg-slate-900 rounded-xl shadow-xl border border-slate-100 dark:border-slate-800 p-1 z-50 animate-in fade-in zoom-in-95 duration-200">
                      <button id="ai-new-chat-btn" class="w-full text-left px-3 py-2 text-sm text-slate-600 dark:text-slate-300 hover:bg-slate-50 dark:hover:bg-slate-800 rounded-lg flex items-center gap-2 transition-colors">
                          <i data-lucide="trash-2" class="w-4 h-4 text-red-500"></i>
                          <span>Clear History & New Chat</span>
                      </button>
                  </div>
              </div>

              <button id="ai-close-btn" class="p-2 text-slate-400 hover:text-slate-600 dark:hover:text-slate-200 hover:bg-slate-100 dark:hover:bg-slate-800 rounded-lg transition-colors">
                <i data-lucide="x" class="w-5 h-5"></i>
              </button>
          </div>
        </div>

        <!-- Messages Area -->
        <div id="ai-chat-messages" class="flex-1 overflow-y-auto p-5 scroll-smooth bg-white dark:bg-slate-900 space-y-5">
          <!-- Messages will act here -->
        </div>

        <!-- Input Area -->
        <div class="p-4 bg-white dark:bg-slate-900 border-t border-slate-100 dark:border-slate-800 shrink-0">
          <div class="relative flex items-end gap-2 bg-slate-50 dark:bg-slate-950 border border-slate-200 dark:border-slate-800 rounded-xl p-2 focus-within:ring-2 focus-within:ring-indigo-500/20 focus-within:border-indigo-500 transition-all">
            <textarea id="ai-input" 
              placeholder="Ask anything about MaungDB..." 
              class="w-full bg-transparent border-none focus:ring-0 text-sm text-slate-700 dark:text-slate-200 placeholder-slate-400 resize-none max-h-32 py-2.5 px-2"
              rows="1"></textarea>
            <button id="ai-send-btn" 
              class="p-2.5 bg-indigo-600 hover:bg-indigo-700 text-white rounded-lg shadow-sm hover:shadow-md transition-all disabled:opacity-50 disabled:cursor-not-allowed">
              <i data-lucide="send-horizontal" class="w-4 h-4"></i>
            </button>
          </div>
          <div class="text-center mt-2">
            <span class="text-[10px] text-slate-400 dark:text-slate-500">Powered by MaungDB AI • Press Enter to send</span>
          </div>
        </div>
      </div>
    `;

    const container = document.createElement('div');
    container.id = 'ai-widget-wrapper';
    container.innerHTML = widgetHTML;
    document.body.appendChild(container);
  }

  toggle() {
    this.isOpen ? this.close() : this.open();
  }

  toggleMenu() {
    const menu = document.getElementById('ai-menu-dropdown');
    menu.classList.toggle('hidden');
  }

  clearHistory() {
    // Clear array
    this.chatHistory = [];
    // Clear localStorage
    localStorage.removeItem('maung_ai_history');
    // Clear UI
    const messagesDiv = document.getElementById('ai-chat-messages');
    messagesDiv.innerHTML = '';
    // Show Welcome Again
    this.showWelcome();
    // Hide menu
    document.getElementById('ai-menu-dropdown').classList.add('hidden');
  }

  open() {
    this.isOpen = true;
    const container = document.getElementById('ai-chat-container');
    container.classList.remove('scale-95', 'opacity-0', 'pointer-events-none');
    container.classList.add('scale-100', 'opacity-100', 'pointer-events-auto');
    document.getElementById('ai-input').focus();
    this.scrollToBottom();
  }

  close() {
    this.isOpen = false;
    const container = document.getElementById('ai-chat-container');
    container.classList.add('scale-95', 'opacity-0', 'pointer-events-none');
    container.classList.remove('scale-100', 'opacity-100', 'pointer-events-auto');
    document.getElementById('ai-menu-dropdown').classList.add('hidden');
  }

  showWelcome() {
    const messagesDiv = document.getElementById('ai-chat-messages');
    messagesDiv.innerHTML = `
      <div class="flex flex-col items-center justify-center py-8 text-center ai-welcome animate-in zoom-in-95 duration-500">
        <div class="w-16 h-16 bg-gradient-to-br from-indigo-500 to-purple-600 rounded-2xl flex items-center justify-center text-white shadow-xl shadow-indigo-500/20 mb-4 transform rotate-3 hover:rotate-6 transition-transform">
          <i data-lucide="bot" class="w-8 h-8"></i>
        </div>
        <h2 class="text-lg font-bold text-slate-800 dark:text-white mb-2">Sampurasun! 🐯</h2>
        <p class="text-sm text-slate-500 dark:text-slate-400 max-w-[260px] leading-relaxed mb-6">
          Nepangkeun, abdi <b>Si Maung</b>. Bade naroskeun naon ngeunaan MaungDB?
        </p>
        
        <div class="grid grid-cols-1 gap-2 w-full max-w-[280px]">
          <button onclick="aiAssistant.askQuestion('Kumaha carana ngadamel tabel?')" 
            class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-slate-800/50 hover:bg-indigo-50 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl text-left transition-all text-xs font-medium text-slate-600 dark:text-slate-300 hover:border-indigo-200 group">
            <div class="p-1.5 bg-white dark:bg-slate-700 rounded-lg shadow-sm group-hover:scale-110 transition-transform">
              <i data-lucide="table" class="w-3.5 h-3.5 text-indigo-500"></i>
            </div>
            <span>Kumaha carana ngadamel tabel?</span>
          </button>
          
          <button onclick="aiAssistant.askQuestion('Bikeun conto insert data')" 
            class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-slate-800/50 hover:bg-indigo-50 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl text-left transition-all text-xs font-medium text-slate-600 dark:text-slate-300 hover:border-indigo-200 group">
            <div class="p-1.5 bg-white dark:bg-slate-700 rounded-lg shadow-sm group-hover:scale-110 transition-transform">
              <i data-lucide="database" class="w-3.5 h-3.5 text-emerald-500"></i>
            </div>
            <span>Conto nambahkeun data (SIMPEN)</span>
          </button>
          
          <button onclick="aiAssistant.askQuestion('Jelaskeun syntax TINGALI')" 
            class="flex items-center gap-3 p-3 bg-slate-50 dark:bg-slate-800/50 hover:bg-indigo-50 dark:hover:bg-slate-800 border border-slate-200 dark:border-slate-700 rounded-xl text-left transition-all text-xs font-medium text-slate-600 dark:text-slate-300 hover:border-indigo-200 group">
            <div class="p-1.5 bg-white dark:bg-slate-700 rounded-lg shadow-sm group-hover:scale-110 transition-transform">
              <i data-lucide="book-open" class="w-3.5 h-3.5 text-amber-500"></i>
            </div>
            <span>Jelaskeun syntax TINGALI</span>
          </button>
        </div>
      </div>
    `;

    this.refreshIcons();
  }

  askQuestion(question) {
    document.getElementById('ai-input').value = question;
    this.sendMessage();
  }

  async sendMessage() {
    const input = document.getElementById('ai-input');
    const message = input.value.trim();

    if (!message || this.isTyping) return;

    input.value = '';
    input.style.height = 'auto';

    // Add User Message
    this.addMessage('user', message);
    this.chatHistory.push({ role: 'user', content: message });
    this.showTyping();

    try {
      const response = await fetch('/ai/chat', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({
          message: message,
          // Kirim 10 message terakhir sebagai context
          history: this.chatHistory.slice(-10),
        }),
      });

      const data = await response.json();
      this.hideTyping();

      if (data.success) {
        this.addMessage('assistant', data.reply);
        this.chatHistory.push({ role: 'assistant', content: data.reply });
        this.saveHistory();
      } else {
        this.showError(data.error || 'Something went wrong');
      }
    } catch (error) {
      this.hideTyping();
      this.showError('Connection error: ' + error.message);
    }
  }

  addMessage(role, content) {
    const messagesDiv = document.getElementById('ai-chat-messages');

    // Hapus welcome jika masih ada
    const welcome = messagesDiv.querySelector('.ai-welcome');
    if (welcome) welcome.remove();

    const isUser = role === 'user';
    const messageDiv = document.createElement('div');
    messageDiv.className = `flex gap-3 ${isUser ? 'flex-row-reverse' : 'flex-row'} animate-in fade-in slide-in-from-bottom-2 duration-300`;

    // Avatar
    const avatarHtml = isUser
      ? `<div class="w-8 h-8 rounded-lg bg-indigo-100 flex items-center justify-center shrink-0 border border-indigo-200"><i data-lucide="user" class="w-4 h-4 text-indigo-600"></i></div>`
      : `<div class="w-8 h-8 rounded-lg bg-slate-900 flex items-center justify-center shrink-0 shadow-md"><i data-lucide="bot" class="w-4 h-4 text-white"></i></div>`;

    // Content Parsing
    let innerContent = isUser ? this.escapeHtml(content) : this.renderMarkdown(content);

    const bubbleClass = isUser
      ? 'bg-indigo-600 text-white rounded-2xl rounded-tr-sm shadow-md shadow-indigo-500/10'
      : 'bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 text-slate-700 dark:text-slate-200 rounded-2xl rounded-tl-sm shadow-sm';

    messageDiv.innerHTML = `
      ${avatarHtml}
      <div class="flex flex-col max-w-[85%] group">
        <div class="px-4 py-3 text-sm leading-relaxed ${bubbleClass}">
          ${innerContent}
        </div>
        ${!isUser ? `
        <div class="flex gap-2 mt-1 opacity-0 group-hover:opacity-100 transition-opacity px-1">
           <button class="text-[10px] text-slate-400 hover:text-indigo-600 flex items-center gap-1" onclick="copyToQuery(this.parentElement.previousElementSibling.textContent)">
             <i data-lucide="copy" class="w-3 h-3"></i> Copy
           </button>
        </div>` : ''}
      </div>
    `;

    messagesDiv.appendChild(messageDiv);
    this.refreshIcons();
    this.scrollToBottom();
  }

  renderMarkdown(text) {
    let html = this.escapeHtml(text);

    // Code blocks
    html = html.replace(/```(\w+)?\n([\s\S]*?)```/g, (match, lang, code) => {
      return `
        <div class="my-3 rounded-lg overflow-hidden border border-slate-200 dark:border-slate-700 bg-slate-50 dark:bg-slate-900">
           <div class="px-3 py-1.5 bg-slate-100 dark:bg-slate-800 border-b border-slate-200 dark:border-slate-700 flex justify-between items-center text-xs text-slate-500 font-mono">
             <span>${lang || 'sql'}</span>
             <button class="hover:text-indigo-600" title="Copy"><i data-lucide="copy" class="w-3 h-3"></i></button>
           </div>
           <div class="p-3 overflow-x-auto">
             <code class="font-mono text-xs text-slate-700 dark:text-slate-300 whitespace-pre">${code.trim()}</code>
           </div>
        </div>`;
    });

    // Inline code
    html = html.replace(/`([^`]+)`/g, '<code class="px-1.5 py-0.5 rounded bg-slate-100 dark:bg-slate-700 text-indigo-600 dark:text-indigo-400 font-mono text-xs border border-slate-200 dark:border-slate-600">$1</code>');
    // Bold
    html = html.replace(/\*\*([^*]+)\*\*/g, '<strong class="font-semibold text-slate-900 dark:text-white">$1</strong>');
    // Newline
    html = html.replace(/\n/g, '<br>');
    // Unescape html for our tags
    html = html.replace(/&lt;/g, '<').replace(/&gt;/g, '>').replace(/&quot;/g, '"').replace(/&#39;/g, "'");

    return html;
  }

  escapeHtml(text) {
    return text.replace(/&/g, "&amp;").replace(/</g, "&lt;").replace(/>/g, "&gt;").replace(/"/g, "&quot;").replace(/'/g, "&#39;");
  }

  showTyping() {
    this.isTyping = true;
    const messagesDiv = document.getElementById('ai-chat-messages');

    const typingHTML = `
      <div id="ai-typing" class="flex gap-3 flex-row animate-in fade-in slide-in-from-bottom-2 duration-300">
        <div class="w-8 h-8 rounded-lg bg-slate-900 flex items-center justify-center shrink-0 shadow-md">
           <i data-lucide="bot" class="w-4 h-4 text-white"></i>
        </div>
        <div class="px-4 py-3 bg-white dark:bg-slate-800 border border-slate-100 dark:border-slate-700 rounded-2xl rounded-tl-sm shadow-sm flex items-center gap-1.5">
           <span class="w-1.5 h-1.5 bg-slate-400 rounded-full animate-bounce"></span>
           <span class="w-1.5 h-1.5 bg-slate-400 rounded-full animate-bounce delay-100"></span>
           <span class="w-1.5 h-1.5 bg-slate-400 rounded-full animate-bounce delay-200"></span>
        </div>
      </div>
    `;

    messagesDiv.insertAdjacentHTML('beforeend', typingHTML);
    this.refreshIcons();
    this.scrollToBottom();
  }

  hideTyping() {
    this.isTyping = false;
    const typing = document.getElementById('ai-typing');
    if (typing) typing.remove();
  }

  showError(msg) {
    const messagesDiv = document.getElementById('ai-chat-messages');
    const errHTML = `
      <div class="flex justify-center animate-in fade-in zoom-in duration-300">
        <div class="bg-red-50 text-red-600 text-xs py-2 px-3 rounded-lg border border-red-100 flex items-center gap-2">
          <i data-lucide="alert-circle" class="w-3.5 h-3.5"></i>
          <span>${msg}</span>
        </div>
      </div>
    `;
    messagesDiv.insertAdjacentHTML('beforeend', errHTML);
    this.refreshIcons();
    this.scrollToBottom();
  }

  scrollToBottom() {
    const messagesDiv = document.getElementById('ai-chat-messages');
    setTimeout(() => { messagesDiv.scrollTop = messagesDiv.scrollHeight; }, 50);
  }

  saveHistory() {
    localStorage.setItem('maung_ai_history', JSON.stringify(this.chatHistory.slice(-20)));
  }

  loadHistory() {
    const saved = localStorage.getItem('maung_ai_history');
    if (saved) {
      this.chatHistory = JSON.parse(saved);
      if (this.chatHistory.length > 0) {
        document.getElementById('ai-chat-messages').innerHTML = '';
        this.chatHistory.forEach(msg => this.addMessage(msg.role, msg.content));
      }
    }
  }
}

// Init
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', () => window.aiAssistant = new AIAssistant());
} else {
  window.aiAssistant = new AIAssistant();
}
