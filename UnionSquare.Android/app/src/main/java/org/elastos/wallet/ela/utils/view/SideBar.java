package org.elastos.wallet.ela.utils.view;

import android.content.Context;
import android.graphics.Canvas;
import android.graphics.Paint;
import android.graphics.Rect;
import android.support.annotation.Nullable;
import android.support.v4.content.ContextCompat;
import android.util.AttributeSet;
import android.view.MotionEvent;
import android.view.View;

import org.elastos.wallet.R;
import org.elastos.wallet.ela.utils.ScreenUtil;


public class SideBar extends View {
    private Paint paint;
    private static final String[] LETTERS = new String[]{
            "A", "B", "C", "D", "E", "F",
            "G", "H", "I", "J", "K", "L",
            "M", "N", "O", "P", "Q", "R",
            "S", "T", "U", "V", "W", "X",
            "Y", "Z"/*, "#"*/
    };
    private float spanHeight;
    private float spanWidth;
    private int startX;
    private int startY;

    private int preItem = -1;

    public SideBar(Context context) {
        this(context, null);
    }

    public SideBar(Context context, @Nullable AttributeSet attrs) {
        this(context, attrs, 0);
    }

    public SideBar(Context context, @Nullable AttributeSet attrs, int defStyleAttr) {
        super(context, attrs, defStyleAttr);
        init();
    }

    private void init() {
        paint = new Paint(Paint.ANTI_ALIAS_FLAG);
        paint.setTextSize(ScreenUtil.dp2px(getContext(), 12));
        paint.setColor(ContextCompat.getColor(getContext(), R.color.whiter));
    }

    @Override
    protected void onDraw(Canvas canvas) {
        super.onDraw(canvas);

        for (int i = 0; i < LETTERS.length; i++) {
            String letter = LETTERS[i];
            Rect rect = new Rect();
            paint.getTextBounds(letter, 0, letter.length(), rect);
            float x = spanWidth * 0.5f - rect.width() * 0.5f;
            float y = spanHeight * 0.5f + rect.height() * 0.5f + i * spanHeight;
            canvas.drawText(LETTERS[i], x, y, paint);
        }
    }

    @Override
    public boolean onTouchEvent(MotionEvent event) {
        switch (event.getAction()) {
            case MotionEvent.ACTION_UP:
                preItem = -1;
                //invalidate();
                break;
            case MotionEvent.ACTION_MOVE:
                getCurrentPoint(event.getY());
                break;
            case MotionEvent.ACTION_DOWN:
                getCurrentPoint(event.getY());
                break;
        }

        return true;
    }

    private void getCurrentPoint(float y) {
        int currentItem = (int) (y / spanHeight);
        if (currentItem != preItem && currentItem < LETTERS.length && currentItem >= 0) {
            String currentLetter = LETTERS[currentItem];
            preItem = currentItem;
            if (onSelectLetterListner != null) {
                onSelectLetterListner.onSelectLetter(currentLetter);
            }
            //invalidate();
        }

    }

    @Override
    protected void onMeasure(int widthMeasureSpec, int heightMeasureSpec) {
        super.onMeasure(widthMeasureSpec, heightMeasureSpec);
        int height = getMeasuredHeight();
        int[] location = new int[2];
        getLocationOnScreen(location);
        startX = location[0];
        startY = location[1];
        spanHeight = height * 1.0f / LETTERS.length;
        spanWidth = getMeasuredWidth();
    }

    public interface OnSelectLetterListner {
        void onSelectLetter(String letter);
    }

    public void setOnSelectLetterListner(OnSelectLetterListner onSelectLetterListner) {
        this.onSelectLetterListner = onSelectLetterListner;
    }

    private OnSelectLetterListner onSelectLetterListner;
}
