class CommentList extends HTMLElement {
    static get observedAttributes() {
        return ['src'];
    }

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.comments = [];
        this.loading = false;
        this.error = null;
    }

    connectedCallback() {
        if (this.hasAttribute('src')) {
            this.fetchComments();
        }
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (name === 'src' && oldValue !== newValue) {
            this.fetchComments();
        }
    }

    async fetchComments() {
        const src = this.getAttribute('src');
        if (!src) return;
        this.loading = true;
        this.error = null;
        this.render();

        try {
            const response = await fetch(src);
            if (!response.ok) throw new Error(`HTTP error: ${response.status}`);
            const data = await response.json();
            this.comments = Array.isArray(data) ? data : [];
            this.loading = false;
            this.render();
        } catch (err) {
            this.error = err.message;
            this.loading = false;
            this.render();
        }
    }

    formatDate(isoString) {
        if (!isoString) return '';
        const date = new Date(isoString);
        return date.toLocaleString();
    }

    render() {
        this.shadowRoot.innerHTML = `
      <style>
        .comment-list {
          font-family: Arial, sans-serif;
          border: 1px solid #ddd;
          border-radius: 6px;
          padding: 1em;
          background: #fafafa;
        }
        .comment {
          border-bottom: 1px solid #eee;
          padding: 0.7em 0;
        }
        .comment:last-child {
          border-bottom: none;
        }
        .author {
          font-weight: bold;
          margin-right: 1em;
        }
        .timestamp {
          color: #888;
          font-size: 0.85em;
        }
        .text {
          margin: 0.2em 0 0 0;
        }
        .error {
          color: red;
        }
      </style>
      <div class="comment-list">
        ${this.loading
                ? `<div>Loading comments...</div>`
                : this.error
                    ? `<div class="error">Error: ${this.error}</div>`
                    : this.comments.length === 0
                        ? `<div>No comments found.</div>`
                        : this.comments.map(comment => `
                  <div class="comment">
                    <span class="author">${comment.AuthorName || 'Unknown Author'}</span>
                    <span class="timestamp">${this.formatDate(comment.Created)}</span>
                    <div class="text">${comment.CommentText || ''}</div>
                  </div>
                `).join('')
            }
      </div>
    `;
    }
}

customElements.define('comment-list', CommentList);
