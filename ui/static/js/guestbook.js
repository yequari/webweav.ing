class GuestbookForm extends HTMLElement {
    static get observedAttributes() {
        return ['guestbook'];
    }

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this._guestbook = this.getAttribute('guestbook') || '';
        this._postUrl = `${this._guestbook}/comments/create/remote`
        this.render();
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (name === 'guestbook') {
            this._guestbook = newValue;
            this.render();
        }
    }

    render() {
        this.shadowRoot.innerHTML = `
      <style>
        form div {
          margin-bottom: 0.75em;
        }
        label {
          display: block;
          font-weight: bold;
          margin-bottom: 0.25em;
        }
        input[type="text"], textarea {
          width: 100%;
          box-sizing: border-box;
        }
        input[type="submit"] {
          font-size: 1em;
          padding: 0.5em 1em;
        }
      </style>
      <form action="${this._postUrl}" method="post">
        <div>
          <label for="authorname">Name</label>
          <input type="text" name="authorname" id="authorname"/>
        </div>
        <div>
          <label for="authoremail">Email (Optional)</label>
          <input type="text" name="authoremail" id="authoremail"/>
        </div>
        <div>
          <label for="authorsite">Site Url (Optional)</label>
          <input type="text" name="authorsite" id="authorsite"/>
        </div>
        <div>
          <label for="content">Comment</label>
          <textarea name="content" id="content"></textarea>
        </div>
        <div>
          <input type="hidden" value="${window.location.pathname}" name="redirect" id="redirect" />
        </div>
        <div>
          <input type="submit" value="Submit"/>
        </div>
      </form>
    `;
    }
}

class CommentList extends HTMLElement {
    static get observedAttributes() {
        return ['guestbook'];
    }

    constructor() {
        super();
        this.attachShadow({ mode: 'open' });
        this.comments = [];
        this.loading = false;
        this.error = null;
    }

    connectedCallback() {
        if (this.hasAttribute('guestbook')) {
            this.fetchComments();
        }
    }

    attributeChangedCallback(name, oldValue, newValue) {
        if (name === 'guestbook' && oldValue !== newValue) {
            this.fetchComments();
        }
    }

    async fetchComments() {
        const guestbook = this.getAttribute('guestbook');
        if (!guestbook) return;
        this.loading = true;
        this.error = null;
        this.render();

        const commentsUrl = `${guestbook}/comments`

        try {
            const response = await fetch(commentsUrl);
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

customElements.define('guestbook-form', GuestbookForm);
customElements.define('guestbook-comments', CommentList);

